## Changes:
* Longer share tokens 8 vs 20
* Wider cards to support longer tokens
* Reload login page on wrong creds. Why? Even if your CF, Proxy, Auth all block access to login, login page can be accessed (internally presented). A forced reload will send a new request, effectively allowing above mentioned options to catch the link and effectively apply the set rules.
* New option:
  - `--publiclogin`: Removes any and all login buttons from the sidebar if set to `false`. Why? To prevent login attempts from a share page.
  - `--publicurl`: Creates shareable url with host/path as base. Explained further in detail below.


## Security Implementation:
If you ask me, best way is to prevent the access to /base_url/api/login. The client will send a post request to this path, which then allows users to authenticate. And thus is the best way to prevent bruteforce attacks.

Below is the recommended nginx.conf
```
worker_processes 1;

events {
    worker_connections 1024;
}

http {
    server_tokens off;
    include mime.types;
    default_type application/octet-stream;
    
    add_header X-Frame-Options "DENY" always;
    add_header Referrer-Policy "no-referrer" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; font-src 'self' data:; img-src 'self' data:; connect-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self';" always;
    
    sendfile on;
    send_timeout 5m;
    proxy_buffering on;
    keepalive_timeout 65;
    proxy_buffers 64 256k;
    proxy_http_version 1.1;
    proxy_read_timeout 360;
    proxy_send_timeout 360;
    proxy_connect_timeout 360;
    proxy_no_cache $cookie_session;
    proxy_redirect http:// $scheme://;
    proxy_cache_bypass $cookie_session;
    proxy_set_header Host $host;
    proxy_set_header Connection "Upgrade";
    proxy_set_header Accept-Encoding gzip;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header X-Forwarded-Ssl on;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $http_host;
    proxy_set_header X-Forwarded-Uri $request_uri;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_buffers 16 8k;
    gzip_min_length 256;
    gzip_disable "msie6";
    gzip_http_version 1.1;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    ssl_session_timeout 5m;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_certificate_key path_to_key;
    ssl_certificate path_to_certificate;
    ssl_ciphers 'TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-CHACHA20-POLY1305-SHA256:!ADH:!AECDH:!MD5:!DSS';

    server {
        listen 999 ssl;
        server_name lostb053.de;
        location = / {
           return 444;
        }
        location /files/static/ {
            proxy_pass http://127.0.0.1:8080/files/static/;
            proxy_intercept_errors on;      
            error_page 404 = @drop;
        }
        location /files/share/ {
            proxy_pass http://127.0.0.1:8080/files/share/;
            proxy_intercept_errors on;      
            error_page 404 = @drop;
        }
        location /files/api/public/share/ {
            proxy_pass http://127.0.0.1:8080/files/api/public/share/;
            proxy_intercept_errors on;      
            error_page 404 = @drop;
        }
        location /api/public/dl/ {
            proxy_pass http://127.0.0.1:8080/files/api/public/dl/;
            proxy_intercept_errors on;      
            error_page 404 = @drop;
        }
        location / {
            proxy_pass http://127.0.0.1:8080/files/api/public/dl/;
            proxy_intercept_errors on;      
            error_page 404 = @drop;
        }
        location @drop {
            return 444;
        }
    }
}
```