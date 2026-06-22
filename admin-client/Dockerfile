# Serve stage directly for static dashboard
FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY index.html /usr/share/nginx/html/index.html
COPY support.js /usr/share/nginx/html/support.js
COPY assets /usr/share/nginx/html/assets
COPY uploads /usr/share/nginx/html/uploads
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
