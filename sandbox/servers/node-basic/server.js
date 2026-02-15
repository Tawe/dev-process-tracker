const http = require('http');

const port = Number(process.env.PORT || 3100);

const server = http.createServer((req, res) => {
  if (req.url === '/health') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ ok: true, service: 'node-basic', port }));
    return;
  }

  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end(`node-basic running on ${port}\n`);
});

server.listen(port, () => {
  console.log(`[node-basic] listening on http://localhost:${port}`);
});
