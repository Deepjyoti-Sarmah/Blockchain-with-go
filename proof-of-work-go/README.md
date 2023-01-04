# Proof Of Work Blockchain

A simple proof of work blockchain using golang

To run -

`go run main.go`

Open `http://localhost:8080` in a browser and you will see one block.

To add blocks, you send a POST request to localhost:8080 using CURL. Send a BPM like {"BPM":75} in the body of this post request.

`curl -X POST -H "Content-Type: application/json" -d '{"name": "75"}' http://localhost:8080`

Your terminal will start performing the work.

Then hit the same terminal from a browser and you will an added block.
