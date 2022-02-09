# yaml-encrypter-decrypter

Утилита для win/linux платформы, позволяющая шифровать в AES значения паролей/секретов в файлах YAML формата

Актуально для тех, кто не использует hashicorp vault, но не хочет хранить секретные данные в git репозитории.

# Как использовать
У утилиты 4 флага, значения которых задано по умолчанию.
```
  -debug string
        режим откладки, выводит в stdout планируемые изменения, но не изменяет yaml файл
        debug mode, print encode/decode to stdout (default "false")
  -env string
        название-начало блока, значения которых надо шифровать
        block-name for encode/decode (default "env:")
  -filename string
        файл,который необходимо зашифровать/дешифровать
        filename for encode/decode (default "values.yaml")
  -key string
        секретный ключ. Обязательно использоание длины в 32 бита.
        после "пилота" будет убрано дефолтное значение
        key for encode, only length 32bit (default "8d9b2dd4c94e8ac7ef742fc0ed162adf49ef8676f906517de1d5085a817ec824")
```

# Варианты запуска утилиты
`go_build_test_go.exe`

`go_build_test_go.exe -key 12345678123456781234567812345678`

`go_build_test_go.exe -filename application.yaml`

`go_build_test_go.exe -filename application.yaml -key 12345678123456781234567812345678`

# Особенности 
Так, как это MVP, есть ряд особенностей:
- есть дефолтный key, после MVP будет убрано дефолтное значение
- ключ пока длинной 32бита, нет поддержки спецсимволов
- от использования библиотек gopkg.in/yaml.v3 и gopkg.in/yaml.v2 пришлось отказаться, потому как они на ходу конвертят в json формат, тем самым затирая комментарии. Задача утилиты шифровать секреты, а не стирать комменты, которые зачастую очень важны.
- запуск утилиты шифрует/дешифрует YAML файл по ключевому значению `AES256-encoded:` в тексте, отдельного флага на декрипт/экрипт нет, задача максимально упростить работу.
- пока сваливается с ошибкой, если значение пустое(не задано), но енкодит, если задано, но пустое ("")
- строчка с комментарием под блоком env: так же шифруется, потом добавлю проверку.

# DEMO

before encrypt:

`./go_build_test_go.exe`
```
#first comment
env:
  rainc: 4354
  coins: 4354
str: #comment
  1: 345343
  2: e5w5g345t
  aerfger:
    rrr: ffgragf
    sd: 4354
    #comment
    env:
      srfgar: 4354

```

after encrypt:

`./go_build_test_go.exe`
```
#first comment
env:
  rainc: AES256-encoded:76bfd42db6f371588ad2a3402130822917f71963d3497339364ff7a242f9cbcd
  coins: AES256-encoded:164d03d0aa62f3af2f9685d64bf07b3d3ada66c527b518aa36051caf0d2b98b3
str: #comment
  1: 345343
  2: e5w5g345t
  aerfger:
    rrr: ffgragf
    sd: 4354
    #comment
    env:
      srfgar: AES256-encoded:035dec914142274f9d2c313adbb1a89176a38cd9116e01a4447acf3205b4120b

```