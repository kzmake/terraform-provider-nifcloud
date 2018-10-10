terraform-provider-nifcloud
==================

Unofficial Terraform Nifcloud provider

Description
------------

Terraform から Nifcloud を操作するためのプラグインです

Requirements
------------

このプロジェクトを実行するには以下が必要です:

- [Terraform](https://www.terraform.io/downloads.html) 0.10+
- [Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)

Install
-------

`$GOPATH/src/github.com/kzmake/terraform-provider-nifcloud` へリポジトリをクローンして、 `go build` を実施しバイナリファイルを作成してください

```sh
$ mkdir -p $GOPATH/src/github.com/kzmake; cd $GOPATH/src/github.com/kzmake
$ git clone https://github.com/kzmake/terraform-provider-nifcloud.git; cd terraform-provider-nifcloud
$ go build
```

ビルドしたバイナリファイルを `~/.terraform.d/plugins/` (Windowsの場合は、`%APPDATA%/terraform.d/plugins/`) に配置してインストール完了です

```sh
$ mkdir -p ~/.terraform.d/plugins/
$ mv terraform-provider-nifcloud ~/.terraform.d/plugins/
```

Terraform のインストールは[公式ページ](https://www.terraform.io/intro/getting-started/install.html)を参考


Preparation
-----------

適当な `*.tf` ファイルを作成してください (ここでは KeyPair の例を示します)

```
provider "nifcloud" {
    access_key = "your access key"
    secret_key = "your secret access key"
    region = "jp-east-1"
}

resource "nifcloud_keypair" "example_ssh_key" {
  name = "nifcloudKey"
  public_key_material = "base64エンコードされた公開鍵の文字列"
  description = "nifcloud_hogehoge"
}
```

Usage
-----

`terraform-provider-nifcloud` をインストール後は、 `terraform init` を実施し、下記のコマンドで `*.tf` を適用します

```sh
$ terraform plan
$ terraform apply
```


Contributing
------------

PR歓迎


Support and Migration
---------------------

特に無し

License
-------

[MIT License](http://petitviolet.mit-license.org/)
