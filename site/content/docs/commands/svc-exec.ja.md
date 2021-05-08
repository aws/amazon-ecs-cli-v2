# svc exec
```
$ copilot svc exec
```

## �R�}���h�̊T�v
`copilot svc exec` �́AService �Ŏ��s���̃R���e�i�ɑ΂��ăR�}���h�����s���܂��B

## �t���O
```
  -a, --app string         Name of the application.
  -c, --command string     Optional. The command that is passed to a running container. (default "/bin/bash")
      --container string   Optional. The specific container you want to exec in. By default the first essential container will be used.
  -e, --env string         Name of the environment.
  -h, --help               help for exec
  -n, --name string        Name of the service, job, or task group.
      --task-id string     Optional. ID of the task you want to exec in.
```

## ���s��

"frontend" Service �̃^�X�N�ɃC���^���N�e�B�u�ȃZ�b�V�������J�n���܂��B

```bash
$ copilot svc exec -a my-app -e test -n frontend
```
"backend" Service ���� ID "8c38184" ����n�܂�^�X�N�� 'ls' �R�}���h�����s���܂��B

```bash
$ copilot svc exec -a my-app -e test --name backend --task-id 8c38184 --command "ls"
```

## �o�͗�

<iframe width="560" height="315" src="https://www.youtube.com/embed/Evrl9Vux31k" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

!!! info
    1. Service �f�v���C�O�� Manifest �� `exec: true` ���ݒ肳��Ă��邱�Ƃ��m�F���Ă��������B
    2. ����ɂ�� Service �� Fargate Platform Version �� 1.4.0 �ɃA�b�v�f�[�g����܂��̂ł����ӂ��������B�v���b�g�t�H�[���o�[�W�������A�b�v�f�[�g����ƁA[ECS �T�[�r�X�̃��v���C�X](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ecs-service.html#cfn-ecs-service-platformversion) �ƂȂ�A�T�[�r�X�̃_�E���^�C�����������܂��B
