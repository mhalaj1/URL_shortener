# URL_shortener

## How to run

Download the src folder (either through CLI or by clicking on the green "Code" button > "Download ZIP" and unzip the src folder).
The src folder should be preferably (although not neccesarily) placed into your Go workspace.

Inside the src/URL_shortener, there is file savedURLs.csv. This file must be empty before the first run of the program. GitHub automatically appends a newline at the end of the file, so you have to manually remove it. Then make sure that the size of the file savedURLs.csv is exactly zero bytes.

Open your terminal and navigate it to the src/URL_shortener folder (it contains main.go).

Build the program using
$ go build main.go

Then run the compiled executable.

## How to use

To shorten URL: open your browser and navigate it to http://localhost:8080/ Then follow on-screen instructions.

To use short URL: simply paste the short URL into your browser's address bar.
