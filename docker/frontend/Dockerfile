FROM nginx:alpine

# Copy static files to nginx html directory
COPY frontend/ /usr/share/nginx/html/

# Copy custom nginx config
COPY docker/frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
