```wireframe state:default viewport:120
{bold}{w}Dev Process Tracker — Health Monitor 0.5.0{/}{/}
{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
{Name (10)       Port    PID      Project         Command                                                                   Health{/}
{m}──────────────  ──────  ───────  ──────────────  ────────────────────────────────────────────────────────────────────────{/}
{sel}▸ {c}api-gateway{/}    8080    45231    gateway         docker run -p 8080:8080 --name api-gateway nginx:latest                    {g}✅{/}{/}
  auth-svc        3000    21847    auth-service    node /home/acme/auth-service/node_modules/.bin/ts-node src/index.ts        {g}✅{/}
  db-migrate      5432    11389    postgres        postgres -D /var/lib/postgresql/16/main -p 5432                           {y}⚠️{/}
  frontend        5173    33912    webapp          node /home/acme/webapp/node_modules/.bin/vite --host                      {g}✅{/}
  redis-cache     6379    8721     cache           redis-server /etc/redis/redis.conf --port 6379                           {g}✅{/}
  scheduler       8901    56290    jobs            python3 -m celery -A tasks worker --loglevel=info -Q scheduler           {r}✘{/}
  search-api      9200    44783    elasticsearch   /usr/share/elasticsearch/bin/elasticsearch -Ehttp.port=9200               {g}✅{/}
  worker-01       5555    29014    background-jobs ruby /home/acme/background-jobs/bin/sidekiq -e development -C sidekiq.yml  {g}✅{/}
  worker-02       5556    29015    background-jobs ruby /home/acme/background-jobs/bin/sidekiq -e development -C sidekiq.yml  {y}⚠️{/}
  ws-relay        4000    17732    realtime        node /home/acme/realtime/node_modules/.bin/ts-node src/websocket.ts       {g}✅{/}
{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
Managed Services (8) ─────────────────────────────────────────── {c}Selected service details{/} ───────────────────────────────
{r}✘ scheduler [crashed]{/}                                            {g}↻{/} {gr}restart{/}   {r}■{/} {gr}stop{/}   {c}✎{/} {gr}edit{/}
{y}■ legacy-proxy [stopped]{/}                                         {g}▶ api-gateway [running]{/}
{g}▶ api-gateway [running]{/}                                           Source: managed
{g}▶ auth-svc [running]{/}                                              PID: 45231
{g}▶ frontend [running]{/}                                              Port: 8080 (tcp)
{g}▶ redis-cache [running]{/}                                           {w}Memory: 128 MB{/}
{g}▶ search-api [running]{/}                                           Cmd: docker run -p 8080:8080 --name api-gateway nginx:latest
{g}▶ ws-relay [running]{/}                                             Dir: /home/acme/gateway
{y}■ staging-fe [stopped]{/}                                            Started: 2026-06-11 09:14:32
{y}⚠ worker-02 [degraded]{/}                                            Health: {g}✅{/} (8ms) HTTP responding in 8ms
{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
{inv} tab{/} switch list  {inv} enter{/} logs/start  {inv} /{/} filter  {inv} ?{/} toggle help  {inv} g{/} group mode
```

```wireframe state:filter-active viewport:120
{bold}{w}Dev Process Tracker — Health Monitor 0.5.0{/}{/}
{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
{Name (10)       Port    PID      Project         Command                                                                   Health{/}
{m}──────────────  ──────  ───────  ──────────     ────────────────────────────────────────────────────────────────────────{/}
{sel}▸ {c}api-gateway{/}    8080    45231    gateway         docker run -p 8080:8080 --name api-gateway nginx:latest                    {g}✅{/}{/}
  auth-svc        3000    21847    auth-service    node /home/acme/auth-service/node_modules/.bin/ts-node src/index.ts        {g}✅{/}
{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
{inv} /{/} {bold}api{/}                                             {dim}3 matches — showing filtered results{/}
{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
Managed Services (8) ─────────────────────────────────────────── {c}Selected service details{/} ───────────────────────────────
{r}✘ scheduler [crashed]{/}                                            {g}↻{/} {gr}restart{/}   {r}■{/} {gr}stop{/}   {c}✎{/} {gr}edit{/}
{y}■ legacy-proxy [stopped]{/}                                         {g}▶ api-gateway [running]{/}
{g}▶ api-gateway [running]{/}                                           Source: managed
{g}▶ auth-svc [running]{/}                                              PID: 45231
{g}▶ frontend [running]{/}                                              Port: 8080 (tcp)
{g}▶ redis-cache [running]{/}                                           {w}Memory: 128 MB{/}
{g}▶ search-api [running]{/}                                           Cmd: docker run -p 8080:8080 --name api-gateway nginx:latest
{g}▶ ws-relay [running]{/}                                             Dir: /home/acme/gateway
{y}■ staging-fe [stopped]{/}                                            Started: 2026-06-11 09:14:32
{y}⚠ worker-02 [degraded]{/}                                            Health: {g}✅{/} (8ms) HTTP responding in 8ms
{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
{inv} tab{/} switch list  {inv} enter{/} logs/start  {inv} /{/} filter  {inv} ?{/} toggle help  {inv} g{/} group mode
```

```wireframe state:error viewport:120
{bold}{w}Dev Process Tracker — Health Monitor 0.5.0{/}{/}
{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
{r}Scan failed: lsof exited with code 1{/}
{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
{r}✘ scheduler [crashed]{/}                                            {c}Selected service details{/}
{g}▶ api-gateway [running]{/}                                           {r}✘ scheduler [crashed]{/}
{g}▶ auth-svc [running]{/}                                              Source: managed
{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
{dim}Error Details:{/}
  Command: lsof -i -P -n
  Exit code: 1
  Stderr: lsof: WARNING: can't stat() fuse.gvfsd-fuse file system

{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}
{inv} r{/} retry scan  {inv} q{/} quit
```

```wireframe state:edit-modal viewport:120
{dim}{m}Dev Process Tracker — Health Monitor 0.5.0{/}{/}
{dim}{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}{/}
{dim}  (managed list dimmed behind the modal){/}
                           ╭─ Edit service ─────────────────────────────────────────────────╮
                           │                                                                │
                           │  {gr}Name{/}     {sel} api-gateway{g}▏{/}{/}                                        │
                           │  {gr}Dir{/}      /home/acme/gateway                                   │
                           │  {gr}Command{/}  docker run -p 8080:8080 --name api-gateway nginx     │
                           │  {gr}Ports{/}    8080                                                 │
                           │                                                                │
                           │  {gr}tab next field   enter save   esc cancel{/}                      │
                           ╰────────────────────────────────────────────────────────────────╯
```

```wireframe state:add-modal viewport:120
{dim}{m}Dev Process Tracker — Health Monitor 0.5.0{/}{/}
{dim}{m}─────────────────────────────────────────────────────────────────────────────────────────────────────{/}{/}
{dim}  (managed list dimmed behind the modal){/}
                           ╭─ Add service ──────────────────────────────────────────────────╮
                           │                                                                │
                           │  {gr}Name{/}     {sel} {g}▏{/}{/}                                                   │
                           │  {gr}Dir{/}                                                           │
                           │  {gr}Command{/}                                                       │
                           │  {gr}Ports{/}                                                         │
                           │                                                                │
                           │  {gr}tab next field   enter save   esc cancel{/}                      │
                           ╰────────────────────────────────────────────────────────────────╯
```
