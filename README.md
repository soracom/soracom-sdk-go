# SORACOM SDK for Go

SORACOM SDK for Go は、株式会社ソラコムの提供する IoT プラットフォーム SORACOM の API を Go のプログラムから呼び出すためのパッケージです。

SORACOM SDK for Go はまだベータ版です。不足している機能はまだたくさんありますし、実装されている機能にも不具合があるかもしれません。また現在提供している機能が予告なく変更されたり削除されたりする可能性もあります。
問題の報告や機能の要望、コードのコントリビュート等、皆様のご協力を歓迎しております。


## インストール

```
$ go get github.com/soracom/soracom-sdk-go
```

Go のインストール、$GOPATH の設定などは事前に行っておいてください。
Go のバージョンは 1.5.1 (arm64/darwin) で動作確認を行いましたが、他のバージョンでも動作すると思います。


## 使用方法

APIClient のインスタンスを作成し、Auth() 関数で認証を行った後、各 API を呼び出します。以下の例では SIM (Subscriber) の一覧を取得して表示しています。

```
package main

import (
    "fmt"
    "github.com/soracom/soracom-sdk-go"
)

func main() {
    ac := soracom.NewAPIClient(nil)
    email := "test@example.com"
    password := "Your password should not be hard-coded here"

    err := ac.Auth(email, password)
    if err != nil {
        fmt.Printf("auth err: %v\n", err.Error())
        return
    }

    subscribers, _, err := ac.ListSubscribers(nil)
    if err != nil {
        fmt.Printf("err: %v\n", err.Error())
        return
    }

    fmt.Printf("%v", subscribers)
}

```

メタデータサービスにも対応しています。
メタデータサービスを利用するには `soracom.NewMetadataClient(nil)` を呼び出して MetadataClient を作成します。

```
package main

import (
	"fmt"
	"github.com/soracom/soracom-sdk-go"
)

func main() {
	mc := soracom.NewMetadataClient(nil)
	sub, err := mc.GetSubscriber()
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	fmt.Printf("subscriber: %v", sub)
}
```

もう少し詳しい使い方は
http://qiita.com/bearmini/items/6e3f66bc0ef846c8d197
を参照してください。
