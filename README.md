# yaml-encrypter-decrypter

Утилита для win/linux платформы, позволяющая шифровать в AES значения паролей/секретов.

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

# Запуск
go_build_test_go.exe
go_build_test_go.exe -key "12345678123456781234567812345678
go_build_test_go.exe -filename application.yaml
