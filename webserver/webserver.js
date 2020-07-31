var http = require('http');
var os = require('os');

var totalrequests = 0;

function getRandomInt(max) {
    return Math.floor(Math.random() * Math.floor(max));
}

http.createServer(async function(request, response) {
    const max = 10;
    const randInt = getRandomInt(max)
    await sleep(100 * randInt);
    totalrequests += 1
    
    response.writeHead(200);
    response.end("Hello! My name is " + os.hostname() + ". I have served "+ totalrequests + " requests so far.\n");
}).listen(8080)

