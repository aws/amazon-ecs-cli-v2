# env show
```bash
$ copilot env show [flags]
```

## �R�}���h�̊T�v
`copilot env show` �́A����� Environment �Ɋւ���ȉ��̂悤�ȏ���\�����܂��B

* Environment �����郊�[�W�����ƃA�J�E���g
* Environment �� _production_ ���ǂ���
* Environment �Ɍ��݃f�v���C����Ă��� Service 
* Environment �Ɋ֘A����^�O

�I�v�V������ `--resources` �t���O��t����� Environment �Ɋ֘A���� AWS ���\�[�X���\������܂��B

## �t���O
```bash
-h, --help          help for show
    --json          Optional. Outputs in JSON format.
-n, --name string   Name of the environment.
    --resources     Optional. Show the resources in your environment.
```
���ʂ��v���O�����Ńp�[�X�������ꍇ `--json` �t���O�𗘗p���邱�Ƃ��ł��܂��B

## ���s��
"test" Environment �Ɋւ������\�����܂��B
```bash
$ copilot env show -n test
```