// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
"use strict";

const aws = require("aws-sdk");

const defaultSleep = function (ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
};

// These are used for test purposes only
let defaultResponseURL;
let waiter;
let sleep = defaultSleep;
let random = Math.random;
let maxAttempts = 10;

/**
 * Upload a CloudFormation response object to S3.
 *
 * @param {object} event the Lambda event payload received by the handler function
 * @param {object} context the Lambda context received by the handler function
 * @param {string} responseStatus the response status, either 'SUCCESS' or 'FAILED'
 * @param {string} physicalResourceId CloudFormation physical resource ID
 * @param {object} [responseData] arbitrary response data object
 * @param {string} [reason] reason for failure, if any, to convey to the user
 * @returns {Promise} Promise that is resolved on success, or rejected on connection error or HTTP error response
 */
let report = function (
  event,
  context,
  responseStatus,
  physicalResourceId,
  responseData,
  reason
) {
  return new Promise((resolve, reject) => {
    const https = require("https");
    const { URL } = require("url");

    var responseBody = JSON.stringify({
      Status: responseStatus,
      Reason: reason,
      PhysicalResourceId: physicalResourceId || context.logStreamName,
      StackId: event.StackId,
      RequestId: event.RequestId,
      LogicalResourceId: event.LogicalResourceId,
      Data: responseData,
    });

    const parsedUrl = new URL(event.ResponseURL || defaultResponseURL);
    const options = {
      hostname: parsedUrl.hostname,
      port: 443,
      path: parsedUrl.pathname + parsedUrl.search,
      method: "PUT",
      headers: {
        "Content-Type": "",
        "Content-Length": responseBody.length,
      },
    };

    https
      .request(options)
      .on("error", reject)
      .on("response", (res) => {
        res.resume();
        if (res.statusCode >= 400) {
          reject(new Error(`Error ${res.statusCode}: ${res.statusMessage}`));
        } else {
          resolve();
        }
      })
      .end(responseBody, "utf8");
  });
};

/**
 * Requests a public certificate from AWS Certificate Manager, using DNS validation.
 * The hosted zone ID must refer to a **public** Route53-managed DNS zone that is authoritative
 * for the suffix of the certificate's Common Name (CN).  For example, if the CN is
 * `*.example.com`, the hosted zone ID must point to a Route 53 zone authoritative
 * for `example.com`.
 *
 * @param {string} requestId the CloudFormation request ID
 * @param {string} domainName the Common Name (CN) field for the requested certificate
 * @param {string} hostedZoneId the Route53 Hosted Zone ID
 * @returns {string} Validated certificate ARN
 */
const requestCertificate = async function (
  requestId,
  domainName,
  subjectAlternativeNames,
  hostedZoneId,
  region
) {
  const crypto = require("crypto");
  const [acm, route53] = clients(region);
  const reqCertResponse = await acm
    .requestCertificate({
      DomainName: domainName,
      SubjectAlternativeNames: subjectAlternativeNames,
      IdempotencyToken: crypto
        .createHash("sha256")
        .update(requestId)
        .digest("hex")
        .substr(0, 32),
      ValidationMethod: "DNS",
    })
    .promise();

  let record;
  for (let attempt = 0; attempt < maxAttempts && !record; attempt++) {
    const { Certificate } = await acm
      .describeCertificate({
        CertificateArn: reqCertResponse.CertificateArn,
      })
      .promise();
    const options = Certificate.DomainValidationOptions || [];

    if (options.length > 0 && options[0].ResourceRecord) {
      record = options[0].ResourceRecord;
    } else {
      // Exponential backoff with jitter based on 200ms base
      // component of backoff fixed to ensure minimum total wait time on
      // slow targets.
      const base = Math.pow(2, attempt);
      await sleep(random() * base * 50 + base * 150);
    }
  }
  if (!record) {
    throw new Error(
      `DescribeCertificate did not contain DomainValidationOptions after ${maxAttempts} tries.`
    );
  }

  console.log(
    `Creating DNS record into zone ${hostedZoneId}: ${record.Name} ${record.Type} ${record.Value}`
  );
  const changeBatch = await updateRecords(
    route53,
    hostedZoneId,
    "UPSERT",
    record.Name,
    record.Type,
    record.Value
  );
  await waitForRecordChange(route53, changeBatch.ChangeInfo.Id);

  await acm
    .waitFor("certificateValidated", {
      // Wait up to 9 minutes and 30 seconds
      $waiter: {
        delay: 30,
        maxAttempts: 19,
      },
      CertificateArn: reqCertResponse.CertificateArn,
    })
    .promise();

  return reqCertResponse.CertificateArn;
};

/**
 * Deletes a certificate from AWS Certificate Manager (ACM) by its ARN.
 * If the certificate does not exist, the function will return normally.
 *
 * @param {string} arn The certificate ARN
 */
const deleteCertificate = async function (arn, region, hostedZoneId) {
  const [acm, route53] = clients(region);
  try {
    console.log(`Waiting for certificate ${arn} to become unused`);

    let inUseByResources;
    let dnsValidationRecord;

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      const { Certificate } = await acm
        .describeCertificate({
          CertificateArn: arn,
        })
        .promise();

      inUseByResources = Certificate.InUseBy || [];
      dnsValidationRecord = Certificate.DomainValidationOptions || [];
      if (inUseByResources.length) {
        // Deleting resources can be quite slow - so just sleep 30 seconds between checks.
        await sleep(30000);
      } else {
        break;
      }
    }

    if (inUseByResources.length) {
      throw new Error(
        `Certificate still in use after checking for ${maxAttempts} attempts.`
      );
    }

    // Fetch the DNS Validation Record and delete it
    if (
      dnsValidationRecord.length > 0 &&
      dnsValidationRecord[0].ResourceRecord
    ) {
      const record = dnsValidationRecord[0].ResourceRecord;

      // Delete this record now that it's not needed
      const changeBatch = await updateRecords(
        route53,
        hostedZoneId,
        "DELETE",
        record.Name,
        record.Type,
        record.Value
      );
      await waitForRecordChange(route53, changeBatch.ChangeInfo.Id);
    }

    await acm
      .deleteCertificate({
        CertificateArn: arn,
      })
      .promise();
  } catch (err) {
    if (err.name !== "ResourceNotFoundException") {
      throw err;
    }
  }
};

const waitForRecordChange = function (route53, changeId) {
  return route53
    .waitFor("resourceRecordSetsChanged", {
      // Wait up to 5 minutes
      $waiter: {
        delay: 30,
        maxAttempts: 10,
      },
      Id: changeId,
    })
    .promise();
};

const updateRecords = function (
  route53,
  hostedZone,
  action,
  recordName,
  recordType,
  recordValue
) {
  return route53
    .changeResourceRecordSets({
      ChangeBatch: {
        Changes: [
          {
            Action: action,
            ResourceRecordSet: {
              Name: recordName,
              Type: recordType,
              TTL: 60,
              ResourceRecords: [
                {
                  Value: recordValue,
                },
              ],
            },
          },
        ],
      },
      HostedZoneId: hostedZone,
    })
    .promise();
};

const clients = function (region) {
  const acm = new aws.ACM({
    region,
  });
  const route53 = new aws.Route53();
  if (waiter) {
    // Used by the test suite, since waiters aren't mockable yet
    route53.waitFor = acm.waitFor = waiter;
  }
  return [acm, route53];
};

/**
 * Main certificate manager handler, invoked by Lambda
 */
exports.certificateRequestHandler = async function (event, context) {
  var responseData = {};
  var physicalResourceId;
  var certificateArn;

  try {
    switch (event.RequestType) {
      case "Create":
      case "Update":
        certificateArn = await requestCertificate(
          event.RequestId,
          event.ResourceProperties.DomainName,
          event.ResourceProperties.SubjectAlternativeNames,
          event.ResourceProperties.HostedZoneId,
          event.ResourceProperties.Region
        );
        responseData.Arn = physicalResourceId = certificateArn;
        break;
      case "Delete":
        physicalResourceId = event.PhysicalResourceId;
        // If the resource didn't create correctly, the physical resource ID won't be the
        // certificate ARN, so don't try to delete it in that case.
        if (physicalResourceId.startsWith("arn:")) {
          await deleteCertificate(
            physicalResourceId,
            event.ResourceProperties.Region,
            event.ResourceProperties.HostedZoneId
          );
        }
        break;
      default:
        throw new Error(`Unsupported request type ${event.RequestType}`);
    }

    await report(event, context, "SUCCESS", physicalResourceId, responseData);
  } catch (err) {
    console.log(`Caught error ${err}.`);
    await report(
      event,
      context,
      "FAILED",
      physicalResourceId,
      null,
      err.message
    );
  }
};

/**
 * @private
 */
exports.withDefaultResponseURL = function (url) {
  defaultResponseURL = url;
};

/**
 * @private
 */
exports.withWaiter = function (w) {
  waiter = w;
};

/**
 * @private
 */
exports.withSleep = function (s) {
  sleep = s;
};

/**
 * @private
 */
exports.reset = function () {
  sleep = defaultSleep;
  random = Math.random;
  waiter = undefined;
  maxAttempts = 10;
};

/**
 * @private
 */
exports.withRandom = function (r) {
  random = r;
};

/**
 * @private
 */
exports.withMaxAttempts = function (ma) {
  maxAttempts = ma;
};
