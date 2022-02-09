# yaml-encrypter-decrypter

Утилита для win/linux платформы, позволяющая шифровать в AES значения паролей/секретов в файлах YAML формата

Шифрование построено на базе AES-256 CBC, который входит в состав функции Helm 3:
- https://helm.sh/docs/chart_template_guide/function_list/#encryptaes
- https://helm.sh/docs/chart_template_guide/function_list/#decryptaes

Утилита особенно актуальна для тех, кто не использует hashicorp vault, 
но не хочет хранить секретные данные в git репозитории.


# Зачем это? есть же Ansible vault!
- полная совместимость с helm 3 версии, функции `decryptAES` и `encryptAES`. YAML файл можно шифровать и расшифровывать как при помощи утилиты, так и шифровать утилитой, но расшифровывать чартом хелма.
- шифруется не весь файл, как в ansible vault, а только значения переменных. Это очень удобно для git history/ pull request
- не требуется дополнительного ПО: python, ansible, ansible-vault и куча dependency. Кроссплатформенность позволяет сделать бинарник хоть для самого крохотного образа alpine или для "обрезанных дистрибутивов линукса"
- работает везде: linux/windows/macos/wsl/gitbash/raspberry, при компиляции можно выбрать любые платформы. Тот же ansible-vault не работает на gitbash.
- открытый исходный код, при желании можно добавить свои фичи, например шифрование не одного файла, а нескольких или даже перебирать файлы в директориях.
- захотелось



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
        секретный ключ
        после "пилота" будет убрано дефолтное значение
        AES key for encrypt/decrypt (default "}tf&Wr+Nt}A9g{s")
```

# Варианты запуска утилиты
- `yed.exe` 
- `yed.exe -key 12345678123456781234567812345678` 
- `yed.exe -filename application.yaml` 
- `yed.exe -filename application.yaml -key 12345678123456781234567812345678` 

# Особенности 
Так, как это MVP, есть ряд особенностей:
- есть дефолтный key, после MVP будет убрано дефолтное значение
- от использования библиотек gopkg.in/yaml.v3 и gopkg.in/yaml.v2 пришлось отказаться, потому как они на ходу конвертят в json формат, тем самым затирая комментарии. Задача утилиты шифровать секреты, а не стирать комменты, которые зачастую очень важны.
- запуск утилиты шифрует/дешифрует YAML файл по ключевому значению `AES256:` в тексте, отдельного флага на декрипт/экрипт нет, задача максимально упростить работу.
- пока сваливается с ошибкой, если значение пустое(не задано), но енкодит, если задано, но пустое ("")
- строчка с комментарием под блоком env: так же шифруется, потом добавлю проверку.

# BUILD
```
set GOARCH=amd64 && set GOOS=linux && go build -o yed main.go
set GOARCH=amd64 && set GOOS=windows && go build -o yed.exe main.go
```

# EXAMPLE 1

before encrypt:

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

`./yed`
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

# EXAMPLE 2
before encrypt:

```
#first comment
env:
  first: 4354
  # пустая строка

  second: 2345625
str: #comment
  1: 345343

  2: e5w5g345t

  aerfger:
    rrr: 2
    login: p9a358htw45uil9pghl945uighwlp45789wty
    #comment

    env:
      password: rio;jkghertuilgyh5uighsa0459tu4
  apps:
    1:
      env:
        key: value
        key1: value2
    2:
      env:
        key3: value3
```

after encrypt:

`./yed`
```
#first comment
env:
  first: AES256-encoded:b5c10458a34f5661358f5048ae69a743ef8d6731a0762fdbdff9251556030c25
  # AES256-encoded:7f64e647a89663f324be281c675554a4bbd9574356f700ae0c768f2b6a6ef20eb286f871a977dacd

  second: 2345625
str: #comment
  1: 345343

  2: e5w5g345t

  aerfger:
    rrr: 2
    login: p9a358htw45uil9pghl945uighwlp45789wty
    #comment

    env:
      password: AES256-encoded:35173190cf1fe17bff5c7a472a22f76ed94b0def6e6ad8025b1d24e6cbc306945702b8d5724c8dec26c6001f8d383879dfd10e32bf0e5064db20a7
  apps:
    1:
      env:
        key: AES256-encoded:c7f7ea8b52c267b3b9350d5e57cd12f5cb5d2a5226b78af3298dd861c09d0a44a2
        key1: AES256-encoded:d0c4f073cc1df8a87a472164445b569cad2609b6257b050e70f230c9d7a8de44818f
    2:
      env:
        key3: AES256-encoded:daaeefc6343c3570d582035cc02a62296945b8d1aa882a2aa1ad46f145749e92d9dd
```