const http = require('http');

const startPort = Number(process.env.PORT || 3200);
const maxAttempts = 10;

function listenWithFallback(port, attempt = 0) {
  const server = http.createServer((req, res) => {
    if (req.url === '/health') {
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ ok: true, service: 'node-port-fallback', port }));
      return;
    }

    res.writeHead(200, { 'Content-Type': 'text/plain' });
    res.end(`node-port-fallback on ${port}\n`);
  });

  server.on('error', (err) => {
    if (err && err.code === 'EADDRINUSE' && attempt < maxAttempts) {
      const next = port + 1;
      console.log(`[node-port-fallback] Port ${port} in use, trying ${next}`);
      listenWithFallback(next, attempt + 1);
      return;
    }

    console.error('[node-port-fallback] failed to bind', err);
    process.exit(1);
  });

  server.listen(port, () => {
    console.log(`[node-port-fallback] listening on http://localhost:${port}`);
  });
}

listenWithFallback(startPort);
