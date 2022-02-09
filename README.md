# yaml-encrypter-decrypter

Утилита для win/linux платформы, позволяющая шифровать в AES значения паролей/секретов в файлах YAML формата

Утилита особенно актуальна для тех, кто не использует hashicorp vault,
но не хочет хранить секретные данные в git репозитории.

Шифрование построено на базе AES-256 CBC, который входит в состав функции Helm 3:
- https://helm.sh/docs/chart_template_guide/function_list/#encryptaes
- https://helm.sh/docs/chart_template_guide/function_list/#decryptaes

Не совместимо с openssl AES-256 CBC.



# Зачем это? Есть же Ansible vault!
- шифруется не весь файл, как в ansible vault, а только значения переменных. Это очень удобно для git history/ pull request
- не требуется дополнительного ПО: python, ansible, ansible-vault и куча dependency. Кроссплатформенность позволяет сделать бинарник хоть для самого крохотного образа alpine или для "обрезанных дистрибутивов линукса"
- работает везде: linux/windows/macos/wsl/gitbash/raspberry, при компиляции можно выбрать любые платформы. Тот же ansible-vault не работает на gitbash.
- открытый исходный код, при желании можно добавить свои фичи, например шифрование не одного файла, а нескольких или даже перебирать файлы в директориях.
- полная совместимость с helm 3 версии, функции `decryptAES` и `encryptAES`. YAML файл можно шифровать и расшифровывать как при помощи утилиты, так и шифровать утилитой, но расшифровывать чартом хелма.
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

# HELM compatibility (TODO, не проверено!)
Пример для встраивания в чарт helm
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: none
data:
  {{- range $key, $value :=  .Values.env -}}
    {{- $key | nindent 2 -}}: {{ encryptAES $aesKey ($value|quote) | printf "AES256:%s" }}
  {{- end }}
```

# BUILD
```
set GOARCH=amd64 && set GOOS=linux && go build -o yed main.go
set GOARCH=amd64 && set GOOS=windows && go build -o yed.exe main.go
```

# EXAMPLE

before encrypt:

```yaml
#first comment
env:
  rainc: 4354
  # comment two
  coins: 4354
str: # 3 comment
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
```yaml
#first comment
env:
  rainc: AES256:RNAavGUfxj2bsQUL1THSwaEXk/hL8xsQNHVSSGFcx78=
  # AES256:UUETCiN57U0cveOWOvFEskmkobCrfmySMsx9OD1hKVg=
  coins: AES256:HtoAvsZQjrsbDyiWMvgmCWF2lqxBGhP4xccROVJWe+o=
str: # 3 comment
  1: 345343

  2: e5w5g345t

  aerfger:
    rrr: ffgragf
    sd: 4354

    #comment

    env:
      srfgar: AES256:uhkboJTlM2wa5VBrgWQ5njwSBVyEQTEXVF89yH/eteI=

```