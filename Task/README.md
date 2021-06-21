1. The solution has been written in golang.
2. Solution currently supports only ETHBTC and BTCUSD symbols. If there is a need to add any symbol just add it to the array on line 18 in main.go. 
3. To deploy the application either create a binary by running ```go build main.go``` and run it on the required host by running ```./main``` or by directly running ```go run main.go```.
4. The server is configured to run on port 9999 but can be changed by editing line 19 in main.go.
5. Navigate to a browser or postman and hit the ```http://localhost:9999/currency/all```, ```http://localhost:9999/currency/ETHBTC``` or ```http://localhost:9999/currency/BTCUSD```.
6. Use of external libraries is limited. For routing purpose github.com/gorilla/mux is used and to handle websocket connection github.com/gorilla/websocket is used.
7. Go modules have been used to manage dependencies.