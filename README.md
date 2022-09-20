# yaml-encrypter-decrypter

Кроссплатформенная утилита, позволяющая шифровать в AES значения паролей/секретов в файлах YAML формата

Утилита особенно актуальна для тех, кто не использует hashicorp vault,
но не хочет хранить секретные данные в git репозитории.

Шифрование построено на базе AES-256 CBC, который входит в состав функции Helm 3:
- https://helm.sh/docs/chart_template_guide/function_list/#encryptaes
- https://helm.sh/docs/chart_template_guide/function_list/#decryptaes

Не совместимо с openssl AES-256 CBC.

# Вариант использования 1
- Разработчик/девопс качает последние изменения из git репозитория.
- Видит, что файл(ы) YAML с данными зашифрованы(есть префикс "AES256:" в значениях).
- Вводит в консоль CMD/GitBash/WSL/etc пароль для расшифровки: set AESKEY="secretpassword"
- Скачивает бинарник, если его нет.
- Расшифровывает файл YAML при помощи команды `./yed -key=$AESKEY` или `./yed -key=$AESKEY -filename=application.yaml`
- Изменяет нужную переменную.
- Зашифровывает файл YAML при помощи команды `./yed -key=$AESKEY` или `./yed -key=$AESKEY -filename=application.yaml`
- git commit, git push

# Вариант использования 2
- Разработчик/девопс качает последние изменения из git репозитория.
- Видит, что файл(ы) YAML с данными зашифрованы(есть префикс "AES256:" в значениях).
- Вводит в консоль CMD/GitBash/WSL/etc пароль для расшифровки: set AESKEY="secretpassword"
- Скачивает бинарник, если его нет.
- Зашифровывает, не меняя файла, одну переменную, например так `./yed -encrypt PLAINTEXT -key $AESKEY`
- Полученный результат вставляет в зашифрованный YAML файл с префиксом AES256:
```yaml
env: 
  key: AES256:<закодированное_значение>
```
- git commit, git push

Во втором случае в git history будет видна только одна изменённая строчка и не надо декодить/енкодить весь файл.


# Зачем это? Есть же Ansible vault и SOPS mozilla!
- шифруется не весь файл, как штатно в ansible vault/SOPS, а только значения переменных определённого блока 
Это очень удобно для git history/ pull request
- не требуется дополнительного ПО: python, ansible, ansible-vault и куча зависимостей. 
- работает везде: linux/windows/macos/wsl/gitbash/raspberry, при компиляции можно выбрать любые платформы. 
Тот же ansible-vault не работает на gitbash.
- полная совместимость с helm 3 версии, функции `decryptAES` и `encryptAES`. 
YAML файл можно шифровать и расшифровывать как при помощи утилиты, так и шифровать утилитой, но расшифровывать чартом хелма.

# Download
https://github.com/kruchkov-alexandr/yaml-encrypter-decrypter/releases/

# Как использовать
У утилиты 6 флагов, значения у 4-ех задано по умолчанию.
```
  -dry-run true/false
        режим откладки, выводит в stdout планируемые изменения, но не изменяет yaml файл
        dry-run mode, print encode/decode to stdout (default "false")
  -env string
        название-начало блока, значения которых надо шифровать
        block-name for encode/decode (default "secret:")
  -filename string
        файл,который необходимо зашифровать/дешифровать
        filename for encode/decode (default "")
  -key string
        секретный ключ
        после "пилота" будет убрано дефолтное значение
        AES key for encrypt/decrypt (default "")
  -operation string
        при выборе операции выбираем decrypt/encrypt
        Available operations: encrypt, decrypt (default "")
  -value string
        при вводе значения в stdout выводится зашифрованое значение
        value to encrypt/decrypt (default "")

```

# Варианты запуска утилиты

- `yed.exe -filename application.yaml -key 12345678123456781234567812345678` 
- `./yed -encrypt PLAINTEXT`
- `./yed -value S5B4ZY2aA1xXBe8HJ8se5sKb/v2J/b7uzOoifpIByzM=  -key SUPERSECRETpassw0000000rd`



# Особенности 
Так, как это MVP, есть ряд особенностей:
- от использования библиотек gopkg.in/yaml.v3 и gopkg.in/yaml.v2 пришлось отказаться, потому как они на ходу конвертят в json формат, 
тем самым затирая комментарии. Задача утилиты шифровать секреты, а не стирать комменты, которые зачастую очень важны.



# HELM compatibility 
Общая идея: 
- бинарник нужен лишь для енкода/декода **локально** у разработчика/девопса
- все values.yaml файлы хранятся с закодированными значениями в git репозитории
- при деплое бинарник yed даже не нужен(не нужно тащить его на gitlab/teamcity агенты)
- расшифровка идёт при помощи нативных функций helm3

Пример для встраивания в чарт helm ниже:

values.yaml
```yaml
# aesKey: мы получаем через helm upgrade --install .... --set aesKey="СЕКРЕТНЫЙ КЛЮЧ"
env:
  key: AES256:11xkAyke8Dx5dQepPSW+VV4FyNUhbcKC3+63+uuFgO8=

```

template\secret.yaml
```yaml
{{- $aesKey := .Values.aesKey }}
apiVersion: v1
kind: Secret
metadata:
  name: example
  namespace: example
  labels:
    app: example
data:
  {{- range $key, $value :=  .Values.env -}}
  {{- if hasPrefix "AES256:" $value -}}
    {{- $key | nindent 2 -}}: {{ ( trimPrefix "AES256:" $value )  | decryptAES $aesKey | b64enc}}
  {{- end }}
  {{- end }}
```

Запуск хелма:
```shell
set SUPERSECRETAESKEY="}tf&Wr+Nt}A9g{s"
helm template RELEASENAME ./CHARTDIRECTORY --values=values.yaml --set aesKey=$SUPERSECRETAESKEY
```

В итоге получаем при генерации манифеста:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example
  namespace: example
  labels:
    app: example
data:
  key: NDM1NA==
```
Если перевести значения из base64, то будет так:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example
  namespace: example
  labels:
    app: example
data:
  key: 4354
```


# Encrypt/Decrypt one value feature
Можно просто шифровать/расшифровать значения, не перезаписывая файл.
Например для того, чтобы не меняя весь файл, зашифровать одну переменную и копипастом вставить в YAML зашифрованный файл.

Пример использования:
```yaml


$ ./yed -value PLAINTEXT -key SUPERSECRETpassw0000000rd
S5B4ZY2aA1xXBe8HJ8se5sKb/v2J/b7uzOoifpIByzM=

$ ./yed -value S5B4ZY2aA1xXBe8HJ8se5sKb/v2J/b7uzOoifpIByzM=  -key SUPERSECRETpassw0000000rd
PLAINTEXT

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
  # comment two
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
