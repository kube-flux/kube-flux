const http = require('http')
const fs = require('fs')
const port = 3000

const server = http.createServer((req, res) => {
    if (req.url === '/') {
        res.writeHead(200, { 'Content-Type': 'text/html' })
        fs.readFile('index.html', (error, data) => {
            if (error) {
                res.writeHead(404)
                res.write('Error: File not Found')
            } else {
                res.write(data)
            }
            res.end()
        })
    }
})

server.on('connection', (socket) => {
    console.log('New Connection...')
    const ip = socket.remoteAddress;
    console.log('Your IP address is' + ip)
})

server.listen(port, (error) => {
    if (error) {
        console.log('Something went wrong', error)
    } else {
        console.log('Server is listening on port ' + port + '...')
    }
})