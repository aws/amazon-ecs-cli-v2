# svc deploy
```bash
$ copilot svc deploy
```

## �R�}���h�̊T�v

`copilot svc deploy` �́A���[�J���̃R�[�h��ݒ�����Ƀf�v���C���܂��B

Service �f�v���C�̎菇�͈ȉ��̒ʂ�ł��B

1. ���[�J���� Dockerfile ���r���h���ăR���e�i�C���[�W���쐬
2. `--tag` �̒l�A�܂��͍ŐV�� git sha (git �f�B���N�g���ō�Ƃ��Ă���ꍇ) �𗘗p���ă^�O�t��
3. �R���e�i�C���[�W�� ECR �Ƀv�b�V��
4. Manifest �t�@�C���ƃA�h�I���� CloudFormation �Ƀp�b�P�[�W
5. ECS �^�X�N��`�ƃT�[�r�X���쐬 / �X�V

## �t���O

```bash
  -e, --env string                     Name of the environment.
  -h, --help                           help for deploy
  -n, --name string                    Name of the service.
      --resource-tags stringToString   Optional. Labels with a key and value separated with commas.
                                       Allows you to categorize resources. (default [])
      --tag string                     Optional. The service's image tag.
```