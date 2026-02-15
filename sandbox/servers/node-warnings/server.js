const http = require("http");

const port = Number(process.env.PORT || 3600);
const warningEveryMs = Number(process.env.WARNING_EVERY_MS || 5000);

const server = http.createServer((req, res) => {
  if (req.url === "/health") {
    res.writeHead(200, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ ok: true, service: "node-warnings", port }));
    return;
  }
  if (req.url === "/warn") {
    console.warn("[node-warnings] warning: simulated runtime warning via /warn");
  }
  res.writeHead(200, { "Content-Type": "text/plain" });
  res.end(`node-warnings running on ${port}\n`);
});

server.listen(port, () => {
  console.log(`[node-warnings] listening on ${port}`);
});

setInterval(() => {
  console.warn("[node-warnings] warning: periodic test warning");
}, warningEveryMs);
