# php runner

php runner is a tcp server compiling php code from client.

## example

first, clone and run php tcp server.

```bash
git clone https://github.com/snowmerak/php-runner.git
cd php-runner
php app.php
```

and run this go tcp client program.

```go
package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

func main() {
	client, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 39710,
	})
	if err != nil {
		panic(err)
	}

	phpCode := `
<?php 
$a = 1;
?>
<h1>Hello World</h1>
<p>paragrapgh</p>
<?php
echo $a . "\n";
?>`

	client.Write([]byte(phpCode))

	data := bytes.NewBuffer(nil)
	buf := [8192]byte{}
	for {
		n, err := client.Read(buf[:])
		if err != nil {
			log.Println(err)
			break
		}
		data.Write(buf[:n])
		if n < 8192 {
			break
		}
	}

	fmt.Println(data.String())
}
```

you can recieve response just like this.

```bash
 
<h1>Hello World</h1> 
<p>paragrapgh</p> 
1

```
