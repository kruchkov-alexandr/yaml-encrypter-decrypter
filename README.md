# YED, yaml-encrypter-decrypter

Cross-platform utility for encrypting/decrypting values of sensitive data in YAML files.

Utility is especially relevant for developers who can't use Hashicorp Vault or SOPS, but not want to store sensitive data in Git repository.

Based on AES-256 CBC encryption, which is part of Helm 3 functions:
- https://helm.sh/docs/chart_template_guide/function_list/#encryptaes
- https://helm.sh/docs/chart_template_guide/function_list/#decryptaes

Not compatible with openssl AES-256 CBC!

# Use case example 1
- Developer/DevOps fetch last changes from Git repository
```
- git pull --all
```
- Developer/DevOps runs command to decrypt all sensitive data in encrypted YAML file
```
yed -decrypt -key SUPERSECRETPASWORD -file values.yaml
```
- Developer/DevOps makes changes in decrypted YAML file, for example add new variable
- Developer/DevOps runs command to encrypt all sensitive data in decrypted YAML file
```
yed -encrypt -key SUPERSECRETPASWORD -file values.yaml
```
- Developer/DevOps commit last own changes to Git repository
```
git commit / git push
```

# Use case example 2
- Developer/DevOps fetch last changes from Git repository
```
- git pull --all
```
- Developer/DevOps runs command to encrypt new one value (or decrypt)
```
yed  -key SUPERSECRETPASWORD -value NEWVALUE
```
- Developer/DevOps copy encrypted value from STDOUT and paste it to encrypted YAML file
- Developer/DevOps commit last own changes to Git repository
```
git commit / git push
```
In the second case, only one line changed and visible in Git history.
There are no sense to decode/encode the whole YAML file.


# But wait? Why not use SOPS, Ansible Vault or Hashicorp Vault?

- available encryption/decryption of one variable without modify the whole file
  - Convenient for git history/ pull request
- without any additional software like Python, Ansible, Ansible-vault and dependencies 
- Cross-platform: linux/windows/macos/wsl/gitbash/raspberry
    - For example Ansible-vault don't executable on git-bash.
- 100% compatible with Helm(version 3+) functions `decryptAES` and `encryptAES`
    - We can decrypt/encrypt with utility and decrypt in helm templates
- 100% free and open source


# Download
https://github.com/kruchkov-alexandr/yaml-encrypter-decrypter/releases/

# How to use
```
There are 6 flags:
  -dry-run boolean
        dry-run mode, print planned encode/decode to stdout (default "false")
  -env string
        block-name for encode/decode (default "secret:")
  -filename string
        filename for encode/decode (default "")
  -key string
        AES key for encrypt/decrypt (default "")
  -operation string
        Available operations: encrypt, decrypt (default "")
  -value string
        value to encrypt/decrypt (default "")
```

# Examples
```
- `yed.exe -filename application.yaml -key 12345678123456781234567812345678 -operation decrypt` 
- `yed -value PLAINTEXT -key 12345678123456781234567812345678`
- `yed -value S5B4ZY2aA1xXBe8HJ8se5sKb/v2J/b7uzOoifpIByzM= -key 12345678123456781234567812345678`
```

# HELM compatibility
- encrypted sensitive data in YAML file stored in Git with prefix `AES256:`
- utility runs on local side only for encrypt/decrypt, no need to copy it on CI/CD
- decryption on CI/Cd use native helm3 functions without any additional software or utilities

Example:
values.yaml
```yaml
# aesKey: get from "helm upgrade --install .... --set aesKey="SUPERSECRETKEY"
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

Run helm deploy:
```shell
set SUPERSECRETAESKEY="1234567890"
helm template RELEASENAME ./CHARTDIRECTORY --values=values.yaml --set aesKey=$SUPERSECRETAESKEY
```

Generated YAML manifest:
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
Base64 decoded:
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
We can encrypt/decrypt one value without modify the whole file.

Example:
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
YAML file before encrypt:
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
YAML file after encrypt:
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
