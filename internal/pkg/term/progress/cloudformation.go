// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package progress

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/aws/copilot-cli/internal/pkg/aws/cloudformation"
	"github.com/aws/copilot-cli/internal/pkg/stream"
	"golang.org/x/sync/errgroup"
)

// StackSubscriber is the interface to subscribe channels to a CloudFormation stack stream event.
type StackSubscriber interface {
	Subscribe() <-chan stream.StackEvent
}

// ECSServiceRendererOpts is optional configuration for a listening ECS service renderer.
type ECSServiceRendererOpts struct {
	Group      *errgroup.Group
	Ctx        context.Context
	RenderOpts RenderOptions
}

// ListeningStackRenderer returns a tree component that listens for CloudFormation
// resource events from a stack mutated with a changeSet until the streamer stops.
func ListeningStackRenderer(streamer StackSubscriber, stackName, description string, changes []Renderer, opts RenderOptions) DynamicRenderer {
	return &dynamicTreeComponent{
		Root:     ListeningResourceRenderer(streamer, stackName, description, opts),
		Children: changes,
	}
}

// ListeningResourceRenderer returns a tab-separated component that listens for
// CloudFormation stack events for a particular resource.
func ListeningResourceRenderer(streamer StackSubscriber, logicalID, description string, opts RenderOptions) DynamicRenderer {
	comp := &regularResourceComponent{
		logicalID:   logicalID,
		description: description,
		statuses:    []stackStatus{notStartedStackStatus},
		stopWatch:   newStopWatch(),
		stream:      streamer.Subscribe(),
		done:        make(chan struct{}),
		padding:     opts.Padding,
		separator:   '\t',
	}
	go comp.Listen()
	return comp
}

// ListeningECSServiceResourceRenderer is a ListeningResourceRenderer for the ECS service cloudformation resource
// and a ListeningRollingUpdateRenderer to render deployments.
func ListeningECSServiceResourceRenderer(streamer StackSubscriber, ecsDescriber stream.ECSServiceDescriber, logicalID, description string, opts ECSServiceRendererOpts) DynamicRenderer {
	g := new(errgroup.Group)
	ctx := context.Background()
	if opts.Group != nil {
		g = opts.Group
	}
	if opts.Ctx != nil {
		ctx = opts.Ctx
	}
	comp := &ecsServiceResourceComponent{
		cfnStream:    streamer.Subscribe(),
		ecsDescriber: ecsDescriber,
		logicalID:    logicalID,

		group:      g,
		ctx:        ctx,
		renderOpts: opts.RenderOpts,

		resourceRenderer: ListeningResourceRenderer(streamer, logicalID, description, opts.RenderOpts),

		done: make(chan struct{}),
	}
	comp.newDeploymentRender = comp.newListeningRollingUpdateRenderer
	go comp.Listen()
	return comp
}

// regularResourceComponent can display a simple CloudFormation stack resource event.
type regularResourceComponent struct {
	logicalID   string        // The LogicalID defined in the template for the resource.
	description string        // The human friendly explanation of the resource.
	statuses    []stackStatus // In-order history of the CloudFormation status of the resource throughout the deployment.
	stopWatch   *stopWatch    // Timer to measure how long the operation takes to complete.

	padding   int  // Leading spaces before rendering the resource.
	separator rune // Character used to separate columns of text.

	stream <-chan stream.StackEvent
	done   chan struct{}
	mu     sync.Mutex
}

// Listen updates the resource's status if a CloudFormation stack resource event is received.
func (c *regularResourceComponent) Listen() {
	for ev := range c.stream {
		if c.logicalID != ev.LogicalResourceID {
			continue
		}
		updateComponentStatus(&c.mu, &c.statuses, ev)
		updateComponentTimer(&c.mu, c.statuses, c.stopWatch)
	}
	close(c.done) // No more events will be processed.
}

// Render prints the resource as a singleLineComponent and returns the number of lines written and the error if any.
func (c *regularResourceComponent) Render(out io.Writer) (numLines int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	components := stackResourceComponents(c.description, c.separator, c.statuses, c.stopWatch, c.padding)
	return renderComponents(out, components)
}

// Done returns a channel that's closed when there are no more events to Listen.
func (c *regularResourceComponent) Done() <-chan struct{} {
	return c.done
}

// ecsServiceResourceComponent can display an ECS service created with CloudFormation.
type ecsServiceResourceComponent struct {
	// Required inputs.
	cfnStream    <-chan stream.StackEvent   // Subscribed stream to initialize the deploymentRenderer.
	ecsDescriber stream.ECSServiceDescriber // Client needed to create an ECSDeploymentStreamer.
	logicalID    string                     // LogicalID for the service.

	// Optional inputs.
	group      *errgroup.Group // Existing group to catch ECSDeploymentStreamer errors.
	ctx        context.Context // Context for the ECSDeploymentStreamer.
	renderOpts RenderOptions

	// Sub-components.
	resourceRenderer   DynamicRenderer
	deploymentRenderer Renderer

	done                chan struct{}
	mu                  sync.Mutex
	newDeploymentRender func(string, time.Time) DynamicRenderer // Overriden in tests.
}

// Listen creates deploymentRenderers if the service is being created, or updated.
// It closes the Done channel if the CFN resource is Done and the deploymentRenderers are also Done.
func (c *ecsServiceResourceComponent) Listen() {
	renderers := []DynamicRenderer{c.resourceRenderer}
	for ev := range c.cfnStream {
		if c.logicalID != ev.LogicalResourceID {
			continue
		}
		if cloudformation.StackStatus(ev.ResourceStatus).UpsertInProgress() {
			if ev.PhysicalResourceID == "" {
				// New service creates receive two "CREATE_IN_PROGRESS" events.
				// The first event doesn't have a service name yet, the second one has.
				continue
			}
			// Start a deployment renderer if a service deployment is happening.
			renderer := c.newDeploymentRender(ev.PhysicalResourceID, ev.Timestamp)
			c.mu.Lock()
			c.deploymentRenderer = renderer
			c.mu.Unlock()
			renderers = append(renderers, renderer)
		}
	}

	// Close the done channel once all the renderers are done listening.
	for _, r := range renderers {
		<-r.Done()
	}
	close(c.done)
}

// Render writes the status of the CloudFormation ECS service resource, followed with details around the
// service deployment if a deployment is happening.
func (c *ecsServiceResourceComponent) Render(out io.Writer) (numLines int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	buf := new(bytes.Buffer)

	nl, err := c.resourceRenderer.Render(buf)
	if err != nil {
		return 0, err
	}
	numLines += nl

	var deploymentRenderer Renderer = &noopComponent{}
	if c.deploymentRenderer != nil {
		deploymentRenderer = c.deploymentRenderer
	}

	nl, err = deploymentRenderer.Render(buf)
	if err != nil {
		return 0, err
	}
	numLines += nl

	if _, err = buf.WriteTo(out); err != nil {
		return 0, err
	}
	return numLines, nil
}

// Done returns a channel that's closed when there are no more events to Listen.
func (c *ecsServiceResourceComponent) Done() <-chan struct{} {
	return c.done
}

func (c *ecsServiceResourceComponent) newListeningRollingUpdateRenderer(serviceARN string, startTime time.Time) DynamicRenderer {
	cluster, service := parseServiceARN(serviceARN)
	streamer := stream.NewECSDeploymentStreamer(c.ecsDescriber, cluster, service, startTime)
	renderer := ListeningRollingUpdateRenderer(streamer, NestedRenderOptions(c.renderOpts))
	c.group.Go(func() error {
		return stream.Stream(c.ctx, streamer)
	})
	return renderer
}

func updateComponentStatus(mu *sync.Mutex, statuses *[]stackStatus, event stream.StackEvent) {
	mu.Lock()
	defer mu.Unlock()

	*statuses = append(*statuses, stackStatus{
		value:  cloudformation.StackStatus(event.ResourceStatus),
		reason: event.ResourceStatusReason,
	})
}

func updateComponentTimer(mu *sync.Mutex, statuses []stackStatus, sw *stopWatch) {
	mu.Lock()
	defer mu.Unlock()

	// There is always at least two elements {notStartedStatus, <new event>}
	curStatus, nextStatus := statuses[len(statuses)-2], statuses[len(statuses)-1]
	switch {
	case nextStatus.value.InProgress():
		// It's possible that CloudFormation sends multiple "CREATE_IN_PROGRESS" events back to back,
		// we don't want to reset the timer then.
		if curStatus.value.InProgress() {
			return
		}
		sw.reset()
		sw.start()
	default:
		if curStatus == notStartedStackStatus {
			// The resource went from [not started] to a finished state immediately.
			// So start the timer and then immediately finish it.
			sw.start()
		}
		sw.stop()
	}
}

func stackResourceComponents(description string, separator rune, statuses []stackStatus, sw *stopWatch, padding int) []Renderer {
	columns := []string{fmt.Sprintf("- %s", description), prettifyLatestStackStatus(statuses), prettifyElapsedTime(sw)}
	components := []Renderer{
		&singleLineComponent{
			Text:    strings.Join(columns, string(separator)),
			Padding: padding,
		},
	}

	for _, failureReason := range failureReasons(statuses) {
		for _, text := range splitByLength(failureReason, maxCellLength) {
			components = append(components, &singleLineComponent{
				Text:    strings.Join([]string{colorFailureReason(text), "", ""}, string(separator)),
				Padding: padding + nestedComponentPadding,
			})
		}
	}
	return components
}
