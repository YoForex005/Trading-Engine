# Deployment Guide

This guide covers deploying the Trading Platform documentation site to various hosting platforms.

## Prerequisites

- Node.js 20+ installed
- Git installed
- Documentation site built successfully (`npm run build`)

## Build the Site

```bash
# Install dependencies
npm install

# Build for production
npm run build

# Test the build locally
npm run serve
```

The build output will be in the `build/` directory.

## Deployment Options

### Option 1: Vercel (Recommended)

Vercel offers the best performance with edge caching and instant deployments.

#### One-Click Deploy

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https://github.com/your-trading-platform/docs)

#### Manual Deploy

```bash
# Install Vercel CLI
npm install -g vercel

# Deploy to Vercel
vercel

# Deploy to production
vercel --prod
```

#### Custom Domain

1. Go to Vercel dashboard
2. Select your project
3. Go to Settings > Domains
4. Add custom domain: `docs.yourtradingplatform.com`
5. Update DNS records:
   ```
   Type: CNAME
   Name: docs
   Value: cname.vercel-dns.com
   ```

#### Environment Variables

Set in Vercel dashboard:
- `ALGOLIA_APP_ID`: Your Algolia app ID
- `ALGOLIA_API_KEY`: Your Algolia API key
- `GA_TRACKING_ID`: Google Analytics tracking ID

### Option 2: Netlify

Netlify provides excellent CI/CD integration and form handling.

#### One-Click Deploy

[![Deploy to Netlify](https://www.netlify.com/img/deploy/button.svg)](https://app.netlify.com/start/deploy?repository=https://github.com/your-trading-platform/docs)

#### Manual Deploy

```bash
# Install Netlify CLI
npm install -g netlify-cli

# Deploy to Netlify
netlify deploy

# Deploy to production
netlify deploy --prod
```

#### Custom Domain

1. Go to Netlify dashboard
2. Select your site
3. Go to Domain settings
4. Add custom domain: `docs.yourtradingplatform.com`
5. Netlify will provide DNS configuration

#### Continuous Deployment

Create `netlify.toml` (already included):
```toml
[build]
  command = "npm run build"
  publish = "build"
```

Push to GitHub, and Netlify will auto-deploy on every commit.

### Option 3: GitHub Pages

Free hosting directly from your GitHub repository.

#### Setup

1. Update `docusaurus.config.ts`:
   ```typescript
   url: 'https://your-org.github.io',
   baseUrl: '/docs/',
   organizationName: 'your-org',
   projectName: 'docs',
   ```

2. Deploy:
   ```bash
   GIT_USER=<your-username> npm run deploy
   ```

This builds and pushes to `gh-pages` branch.

#### Custom Domain

1. Add `CNAME` file to `static/` directory:
   ```
   docs.yourtradingplatform.com
   ```

2. Configure DNS:
   ```
   Type: CNAME
   Name: docs
   Value: your-org.github.io
   ```

### Option 4: AWS S3 + CloudFront

For enterprise deployments with full AWS integration.

#### Setup S3 Bucket

```bash
aws s3 mb s3://docs.yourtradingplatform.com
aws s3 website s3://docs.yourtradingplatform.com \
  --index-document index.html \
  --error-document 404.html
```

#### Upload Build

```bash
npm run build
aws s3 sync build/ s3://docs.yourtradingplatform.com \
  --delete \
  --cache-control max-age=31536000
```

#### CloudFront Distribution

1. Create CloudFront distribution
2. Origin: S3 bucket
3. Default Root Object: `index.html`
4. Error Pages: Custom 404 → `/404.html`
5. SSL Certificate: Use ACM
6. Custom Domain: `docs.yourtradingplatform.com`

#### Automated Deploy Script

```bash
#!/bin/bash
npm run build
aws s3 sync build/ s3://docs.yourtradingplatform.com --delete
aws cloudfront create-invalidation --distribution-id YOUR_DIST_ID --paths "/*"
```

### Option 5: Docker Deployment

For self-hosted deployments.

#### Create Dockerfile

```dockerfile
FROM node:20-alpine AS build

WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### Build and Run

```bash
# Build Docker image
docker build -t trading-docs .

# Run container
docker run -d -p 80:80 trading-docs
```

#### Deploy to Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: trading-docs
spec:
  replicas: 3
  selector:
    matchLabels:
      app: trading-docs
  template:
    metadata:
      labels:
        app: trading-docs
    spec:
      containers:
      - name: trading-docs
        image: your-registry/trading-docs:latest
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: trading-docs
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: trading-docs
```

## Performance Optimization

### Enable CDN

All major platforms provide CDN automatically:
- Vercel: Vercel Edge Network
- Netlify: Netlify CDN
- AWS: CloudFront
- Cloudflare: Add as DNS proxy

### Compress Assets

Build script already includes compression. Verify:
```bash
# Check if gzip enabled
curl -H "Accept-Encoding: gzip" -I https://docs.yourtradingplatform.com
```

Should see:
```
content-encoding: gzip
```

### Image Optimization

Images are automatically optimized during build:
- WebP format for supported browsers
- Lazy loading enabled
- Responsive image sizes

### Caching Strategy

Configure cache headers:

```nginx
# nginx.conf
location /assets/ {
  expires 1y;
  add_header Cache-Control "public, immutable";
}

location / {
  expires 1h;
  add_header Cache-Control "public, must-revalidate";
}
```

## Search Configuration

### Algolia DocSearch

1. Apply for free DocSearch: https://docsearch.algolia.com/apply
2. Receive Algolia credentials
3. Update `docusaurus.config.ts`:
   ```typescript
   algolia: {
     appId: 'YOUR_APP_ID',
     apiKey: 'YOUR_API_KEY',
     indexName: 'trading-platform',
   }
   ```

### Self-Hosted Search

Alternative to Algolia:

```bash
# Install local search plugin
npm install --save @easyops-cn/docusaurus-search-local

# Update docusaurus.config.ts
plugins: [
  [
    require.resolve("@easyops-cn/docusaurus-search-local"),
    {
      hashed: true,
    },
  ],
],
```

## Analytics Setup

### Google Analytics

Already configured in `docusaurus.config.ts`. Just update tracking ID:
```typescript
gtag: {
  trackingID: 'G-XXXXXXXXXX',
}
```

### Plausible Analytics

Privacy-friendly alternative:
```bash
npm install --save docusaurus-plugin-plausible

# docusaurus.config.ts
plugins: [
  ['docusaurus-plugin-plausible', { domain: 'docs.yourtradingplatform.com' }],
],
```

## Monitoring

### Uptime Monitoring

Use services like:
- UptimeRobot (free)
- Pingdom
- StatusCake

Monitor: `https://docs.yourtradingplatform.com`

### Performance Monitoring

Use Google Lighthouse:
```bash
npm install -g lighthouse
lighthouse https://docs.yourtradingplatform.com --view
```

Target scores:
- Performance: >90
- Accessibility: 100
- Best Practices: 100
- SEO: 100

### Error Tracking

Add Sentry for error tracking:
```bash
npm install --save @docusaurus/plugin-client-redirects

# docusaurus.config.ts
plugins: [
  ['@docusaurus/plugin-client-redirects', {
    fromExtensions: ['html', 'htm'],
  }],
],
```

## SSL/TLS Configuration

All platforms provide free SSL:
- Vercel: Automatic SSL
- Netlify: Let's Encrypt
- GitHub Pages: Automatic SSL
- AWS: Use ACM (AWS Certificate Manager)

### Force HTTPS

Configure in platform settings or add redirect:

```nginx
# nginx
if ($scheme != "https") {
  return 301 https://$host$request_uri;
}
```

## Backup Strategy

### Git-Based Backups

Your documentation is version-controlled:
```bash
git push origin main  # Primary remote
git push backup main  # Backup remote
```

### Automated Backups

GitHub Actions workflow:
```yaml
name: Backup
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly
jobs:
  backup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Backup to S3
        run: |
          aws s3 sync . s3://docs-backup/$(date +%Y-%m-%d)/
```

## CI/CD Pipeline

### GitHub Actions

`.github/workflows/deploy.yml`:
```yaml
name: Deploy Documentation
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-node@v2
        with:
          node-version: '20'
      - run: npm ci
      - run: npm run build
      - uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./build
```

## Troubleshooting

### Build Fails

```bash
# Clear cache
npm run clear

# Clean install
rm -rf node_modules package-lock.json
npm install

# Try again
npm run build
```

### 404 Errors After Deploy

Check:
1. Base URL in config matches deployment
2. Routes are properly configured
3. `404.html` exists in build output

### Slow Build Times

```bash
# Use production build
NODE_ENV=production npm run build

# Parallel processing
npm run build -- --workers 4
```

## Maintenance

### Regular Updates

```bash
# Update dependencies monthly
npm update

# Check for security issues
npm audit

# Fix vulnerabilities
npm audit fix
```

### Content Updates

1. Create feature branch
2. Make changes
3. Test locally: `npm start`
4. Build: `npm run build`
5. Create PR
6. Merge to main → Auto-deploy

## Cost Estimates

### Free Tier Options
- Vercel: Free for personal/commercial
- Netlify: 100GB bandwidth/month free
- GitHub Pages: Free (public repos)

### Paid Options
- Vercel Pro: $20/month (better performance)
- Netlify Pro: $19/month (enhanced features)
- AWS: ~$5-50/month (usage-based)

## Support

Need help with deployment?
- Email: devops@yourtradingplatform.com
- Discord: https://discord.gg/yourtradingplatform
- Docs: https://docusaurus.io/docs/deployment
