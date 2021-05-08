# env delete
```bash
$ copilot env delete [flags]
```

## �R�}���h�̊T�v
`copilot env delete` �́AApplication ���� Environment ���폜���܂��B Environment ���Ɏ��s���̃A�v���P�[�V����������ꍇ�́A�͂��߂� [`copilot svc delete`](../commands/svc-delete.md) �����s����K�v������܂��B

����ɓ�������AEnvironment �p�� AWS CloudFormation �X�^�b�N���폜���ꂽ���Ƃ��m�F���Ă��������B

## �t���O
```
-h, --help             help for delete
-n, --name string      Name of the environment.
    --yes              Skips confirmation prompt.
-a, --app string       Name of the application.
```

## ���s��
"test" Environment ���폜���܂��B
```bash
$ copilot env delete --name test 
```
"test" Environment ���v�����v�g�Ȃ��ō폜���܂��B
```bash
$ copilot env delete --name test --yes
```
