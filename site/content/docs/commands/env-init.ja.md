# env init
```bash
$ copilot env init [flags]
```

## �R�}���h�̊T�v
`copilot env init` �́AService ���s�p�ɐV���� [Environment](../concepts/environments.md) ���쐬���܂��B

����ɓ�����ƁACLI �� VPC, Application Load Balancer, ECS Cluster �Ȃǂ� Service �ŋ��L����鋤�ʂ̃C���t���X�g���N�`���[���쐬���܂��B����ɁA�f�t�H���g�̃��\�[�X�ݒ��������\�[�X�̃C���|�[�g�ȂǁA[Copilot �� Environment ���J�X�^�}�C�Y](../concepts/environments.md#customize-your-environment) �ł��܂��B

[���O�t���v���t�@�C��](../credentials.md#environment-credentials) ���g�p���āAEnvironment ���쐬���� AWS �A�J�E���g�ƃ��[�W�������w�肵�܂��B

## �t���O
AWS Copilot CLI �̑S�ẴR�}���h���l�A�K�{�t���O���ȗ������ꍇ�ɂ͂����̏��̓��͂��C���^���N�e�B�u�ɋ��߂��܂��B�K�{�t���O�𖾎��I�ɓn���ăR�}���h�����s���邱�Ƃł�����X�L�b�v�ł��܂��B
```
Common Flags
      --aws-access-key-id string       Optional. An AWS access key.
      --aws-secret-access-key string   Optional. An AWS secret access key.
      --aws-session-token string       Optional. An AWS session token for temporary credentials.
      --default-config                 Optional. Skip prompting and use default environment configuration.
  -n, --name string                    Name of the environment.
      --prod                           If the environment contains production services.
      --profile string                 Name of the profile.
      --region string                  Optional. An AWS region where the environment will be created.

Import Existing Resources Flags
      --import-private-subnets strings   Optional. Use existing private subnet IDs.
      --import-public-subnets strings    Optional. Use existing public subnet IDs.
      --import-vpc-id string             Optional. Use an existing VPC ID.

Configure Default Resources Flags
      --override-private-cidrs strings   Optional. CIDR to use for private subnets (default 10.0.2.0/24,10.0.3.0/24).
      --override-public-cidrs strings    Optional. CIDR to use for public subnets (default 10.0.0.0/24,10.0.1.0/24).
      --override-vpc-cidr ipNet          Optional. Global CIDR to use for VPC (default 10.0.0.0/16).

Global Flags
  -a, --app string   Name of the application.
```

## ���s��
AWS �v���t�@�C���� "default" �ɁA�f�t�H���g�ݒ���g�p���� test Environment ���쐬���܂��B

```bash
$ copilot env init --name test --profile default --default-config
```

AWS �v���t�@�C���� "prod-admin" �𗘗p���Ċ����� VPC �� prod-iad Environment ���쐬���܂��B
```bash
$ copilot env init --name prod-iad --profile prod-admin --prod \
--import-vpc-id vpc-099c32d2b98cdcf47 \
--import-public-subnets subnet-013e8b691862966cf,subnet-014661ebb7ab8681a \
--import-private-subnets subnet-055fafef48fb3c547,subnet-00c9e76f288363e7f
```

## �o�͗�
![Running copilot env init](https://raw.githubusercontent.com/kohidave/copilot-demos/master/env-init.svg?sanitize=true)
