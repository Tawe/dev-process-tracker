const http = require("http");

const port = Number(process.env.PORT || 3500);
const crashAfterMs = Number(process.env.CRASH_AFTER_MS || 2500);

const server = http.createServer((req, res) => {
  if (req.url === "/health") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ ok: true, service: "node-crash", port }));
    return;
  }
  res.writeHead(200, { "Content-Type": "text/plain" });
  res.end(`node-crash running on ${port}\n`);
});

server.listen(port, () => {
  console.log(`[node-crash] listening on ${port}`);
  console.warn(`[node-crash] warning: crash scheduled in ${crashAfterMs}ms`);
  setTimeout(() => {
    console.error("[node-crash] fatal: intentional crash for testing");
    process.exit(1);
  }, crashAfterMs);
});
