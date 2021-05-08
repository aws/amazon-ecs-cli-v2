# svc init
```bash
$ copilot svc init
```

## �R�}���h�̊T�v

`copilot svc init` �́A�R�[�h�����s���邽�߂ɐV���� [Service](../concepts/services.md) ���쐬���܂��B 

�R�}���h�����s����ƁA CLI �̓��[�J���� `copilot` �f�B���N�g���� Application ���̃T�u�f�B���N�g�����쐬���A������ [Manifest �t�@�C��](../manifest/overview.md) ���쐬���܂��B���R�� Manifest �t�@�C�����X�V���AService �̃f�t�H���g�ݒ��ύX�ł��܂��B�܂� CLI �͑S�Ă� [Environment](../concepts/environments.md) ����v���\�ɂ���|���V�[������ ECR ���|�W�g�����Z�b�g�A�b�v���܂��B

������ Service �� CLI ����g���b�N���邽�� AWS System Manager Parameter Store �ɓo�^����܂��B

���̌�A���ɃZ�b�g�A�b�v���ꂽ Environment ������ꍇ�� `copilot deploy` �����s���� Service ���f�v���C�ł��܂��B

## �t���O

```bash
Required Flags
  -d, --dockerfile string   Path to the Dockerfile.
  -n, --name string         Name of the service.
  -t, --svc-type string     Type of service to create. Must be one of:
                            "Load Balanced Web Service", "Backend Service"

Load Balanced Web Service Flags
      --port uint16   Optional. The port on which your service listens.

Backend Service Flags
      --port uint16   Optional. The port on which your service listens.
```

�e Service type �ɂ͋��ʂ̕K�{�t���O�̑��ɁA�Ǝ��̃I�v�V�����t���O�ƕK�{�t���O������܂��B"frontend" �Ƃ��� Load Balanced Web Service ���쐬����ɂ́A���̂悤�Ɏ��s���܂�


`$ copilot svc init --name frontend --app-type "Load Balanced Web Service" --dockerfile ./frontend/Dockerfile`

## �o�͗�

![Running copilot svc init](https://raw.githubusercontent.com/kohidave/copilot-demos/master/svc-init.svg?sanitize=true)
