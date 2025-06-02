Leverage File Browser (GitHub URL: https://github.com/filebrowser/filebrowser) as the storage & HTTP API layer.
GitHub

Scaffold a React front-end that uses react-dropzone for drag-and-drop and Axios for uploads.
react-dropzone.js.org
Stack Overflow

Wire both services in a single docker-compose.yml; mount a named volume so anything dropped into File Browser instantly appears in the Library tab.
Docker Documentation

Expose File Browser’s REST endpoints (/api/files/:path, /api/raw/:path, /api/zip) to the React UI so you never touch the disk directly.
File Browser
GitHub

With that architecture, you can build, test and hand off a fully working micro-service before folding the code back into the larger project.

Stage 0 – Prerequisites & local toolchain
Install Docker Engine ≥ 24.x and Docker Compose v2 on every dev box.

Verify: docker --version and docker compose version.
Docker Documentation

Install Node 18 LTS and Yarn (or npm ≥ 10) for front-end builds.

Confirm ports 8080 (File Browser) and 5173 (Vite dev server) or 3000 (CRA dev server) are free.

Stage 1 – Project skeleton
text
Copy
Edit
podcaster-uploader/
├─ docker-compose.yml
├─ .env
├─ data/            ← will be bound to the named volume
└─ frontend/
Create the directory early; Compose will create the named volume but it helps to have a local data/ folder for inspection.
Docker Documentation

Stage 2 – Provision File Browser back-end
Add the service to docker-compose.yml:

yaml
Copy
Edit
services:
  filebrowser:
    image: filebrowser/filebrowser:latest     # official image on Docker Hub
    container_name: fb
    volumes:
      - pod_data:/srv                         # named volume
    environment:
      - PUID=1000                             # host UID (non-root)
      - PGID=1000                             # host GID
    ports:
      - 8080:80                               # UI & API
    restart: unless-stopped

volumes:
  pod_data:                                   # created automatically
Image details and environment flags come straight from the Docker Hub page.
Docker Hub

Run once in interactive mode to create the default admin user if you want non-public access:
docker exec -it fb filebrowser users add admin password --perm.admin
File Browser

Optional CORS: add --cors all to File Browser’s command if your React UI lives on a different host or port.
File Browser

Stage 3 – Scaffold the React front-end
Bootstrap with Vite or CRA (Vite is faster; substitute CRA if preferred):

bash
Copy
Edit
cd frontend
npm create vite@latest uploader --template react
cd uploader
npm install
Docker’s own blog shows an identical scaffold being dockerised later.
Docker

Add dependencies:

bash
Copy
Edit
npm i react-dropzone axios
react-dropzone supplies the drag-and-drop hooks; Axios handles multipart/form-data posts.
react-dropzone.js.org
Stack Overflow

Stage 4 – Implement the Upload tab
Create an <UploadPage> component that calls useDropzone(); accept audio/*.

On onDrop, build FormData:

js
Copy
Edit
const form = new FormData();
form.append('files', acceptedFile, acceptedFile.name);
await axios.post('/api/upload', form, { baseURL: import.meta.env.VITE_API });
Axios example follows the widely accepted pattern in community Q&A.
Stack Overflow

Point VITE_API to http://localhost:8080 in .env files for development; use the service name http://filebrowser when the React container runs inside Compose.

File Browser’s upload endpoint is POST /api/upload/:path?override=false; default path is root (/).
GitHub

Stage 5 – Implement the Library tab
Fetch directory listing with GET /api/files/:path (JSON).
File Browser

Render results in a table or grid (file name, size, modified date).

Each row’s download link hits GET /api/raw/:path for direct download/stream.
File Browser

For bulk downloads, File Browser supports GET /api/zip?files=/foo.mp3&files=/bar.wav.
File Browser

Because both Upload and Library talk only to the REST API, any file that already lives in /srv (the mounted volume) appears automatically—no extra code required.

Stage 6 – Containerise the front-end
Add a Dockerfile inside frontend/uploader/:

Dockerfile

# builder
FROM node:18-alpine AS build
WORKDIR /app
COPY . .
RUN npm ci && npm run build

# runner
FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
A multi-stage build keeps the final image tiny; the Nginx pattern mirrors the Docker guide.
Docker

Augment docker-compose.yml:


  frontend:
    build: ./frontend/uploader
    container_name: ui
    environment:
      - VITE_API=http://filebrowser                # internal host
    depends_on:
      - filebrowser
    ports:
      - 3000:80                                    # prod port
Stage 7 – Build & run
bash
Copy
Edit
docker compose build
docker compose up -d
Compose creates the pod_data volume only once; subsequent up commands reuse it, so stored audio survives container restarts.
Docker Documentation

Visit http://localhost:3000 → Upload tab → drag an .mp3.
Switch to Library → the file appears instantly; clicking downloads/streams it through File Browser’s /raw endpoint.

Stage 8 – Smoke tests & validation
CLI test:
curl -F "files=@sample.wav" http://localhost:8080/api/upload → expect {"event":"upload success"}.
GitHub

Large-file retry: pause network mid-upload; File Browser resumes automatically (uses TUS-like chunking).
File Browser

MIME filter: ensure React limits to audio/*; File Browser does server-side checks too.
File Browser

Stage 9 – Hardening & production tweaks
Concern	Mitigation
Auth	Create non-admin File Browser users; pass a JWT or session cookie from your SSO gateway.
File Browser
HTTPS	Terminate TLS at your reverse-proxy (Traefik, Nginx, Caddy).
Back-ups	Use docker run --rm -v pod_data:/data busybox tar czf /backup/pod_data.tgz /data.
Config as code	Add filebrowser.json (users, branding) to the volume; File Browser consumes it at startup.
GitHub

Stage 10 – Integrate into the main project
Mount the same named volume in your legacy application containers; they can read the exact files File Browser writes.

Replace direct file-system access in the monolith with HTTP calls to /api/raw/:path if you’d rather not mount the volume.

CI/CD: build front-end and image in your existing pipeline; Compose file can be templated via Helm/Kustomize when you migrate to Kubernetes later.

Reference link requested
File Browser GitHub: https://github.com/filebrowser/filebrowser
GitHub

Key sources consulted (alphabetical by domain)
Docker Compose volume docs – volumes & reuse.
Docker Documentation

Docker Hub – official filebrowser/filebrowser image.
Docker Hub

File Browser project README & docs – endpoints, flags, CORS.
GitHub
File Browser

GitHub issue thread – confirmed upload API pattern.
GitHub

Medium / Docker blog – multi-stage React image build.
Docker

NPM & official site – react-dropzone usage.
react-dropzone.js.org
npm

RapidAPI / StackOverflow – Axios file-upload snippet.
Stack Overflow
RapidAPI

Reddit – practical explanation of named volumes.
Reddit

Docker Hub (alternative maintainer) – image size & variants.
Docker Hub

GitHub helper CLI for scripted uploads (optional).
GitHub

Hand this guide to the team, and they’ll have everything needed—from initial scaffold to production hardening—to deliver a clean, self-contained Upload + Library module that can be dropped into (or run alongside) your main application stack.
