<?php
$server = stream_socket_server("tcp://0.0.0.0:39710", $errno, $errstr);

if ($server === false) {
    throw new Exception("Could not create server: $errstr ($errno)");
}

for (;;) {
    $client = stream_socket_accept($server);
    if ($client === false) {
        echo "Could not accept client\n";
        continue;
    }

    $request = '';
    while (true) {
        $buf = fread($client, 8192);
        $request .= $buf;
        if (strlen($buf) < 8192) {
            break;
        }
    }

    $php_mode = false;
    $php_code = "";

    $lines = explode("\n", $request);

    ob_start();

    foreach ($lines as $line) {
        $echoed = false;
        foreach (explode(" ", $line) as $word) {
            switch ($php_mode) {
                case true:
                    if ($word == "?>") {
                        eval($php_code);
                        $php_code = "";
                        $php_mode = false;
                    } else {
                        $php_code .= $word . " ";
                    }
                    break;
                case false:
                    if (strcmp($word, '<?php') == 0) {
                        $php_mode = true;
                    } else {
                        echo $word . " ";
                        $echoed = true;
                    }
                    break;
            }
        }
        if ($echoed) {
            echo "\n";
        }
    }

    $response = ob_get_contents();
    ob_end_clean();

    if (fwrite($client, $response) === false) {
        echo "Could not write response\n";
    }

    fclose($client);
}

?>