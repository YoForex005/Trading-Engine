# Quick Start - Documentation Site

Get your Trading Platform documentation site running in 2 minutes.

## Installation

```bash
cd docs-site
npm install
```

## Development

```bash
# Start development server (http://localhost:3000)
npm start

# Start with custom host (accessible on network)
npm run start:host

# The site will auto-reload when you save changes
```

## Building

```bash
# Build for production
npm run build

# Test production build locally
npm run serve:build
```

## Deployment

### Quick Deploy to Vercel (Recommended)

```bash
# Install Vercel CLI
npm install -g vercel

# Deploy
vercel

# Deploy to production
vercel --prod
```

### Quick Deploy to Netlify

```bash
# Install Netlify CLI
npm install -g netlify-cli

# Deploy
netlify deploy

# Deploy to production
netlify deploy --prod
```

### GitHub Pages

```bash
# Update docusaurus.config.ts with your GitHub info
# Then run:
GIT_USER=<your-username> npm run deploy
```

## File Structure

```
docs/                   # Your documentation content (Markdown)
├── getting-started/   # Getting started guides
├── trading-guide/     # Trading tutorials
├── api/              # API documentation
├── admin/            # Admin guides
├── integrations/     # Integration guides
├── reference/        # Reference docs
├── troubleshooting/  # Help & FAQ
└── legal/            # Legal documents

src/                   # Custom React components
static/                # Static files (images, downloads)
blog/                  # Blog posts (optional)
```

## Creating Documentation

### Add a New Page

1. Create a markdown file in `docs/`:
```bash
# Example: docs/getting-started/new-guide.md
```

2. Add frontmatter:
```markdown
---
id: new-guide
title: New Guide
sidebar_label: New Guide
sidebar_position: 10
---

# New Guide Content

Your content here...
```

3. Update `sidebars.ts` to include your new page:
```typescript
gettingStartedSidebar: [
  {
    type: 'category',
    label: 'Getting Started',
    items: [
      'getting-started/introduction',
      'getting-started/new-guide',  // Add here
    ],
  },
],
```

### Add Code Examples

Use fenced code blocks with language specification:

````markdown
```javascript
const client = new TradingClient({
  apiKey: 'YOUR_API_KEY'
});

const balance = await client.getBalance();
console.log(`Balance: $${balance.balance}`);
```
````

### Add Admonitions

```markdown
:::tip Pro Tip
Use this for helpful tips and best practices.
:::

:::warning Important
Use this for warnings and important information.
:::

:::danger Risk Warning
Use this for critical warnings and risks.
:::

:::info Note
Use this for general information.
:::
```

## Customization

### Update Site Config

Edit `docusaurus.config.ts`:
```typescript
const config: Config = {
  title: 'Your Platform',           // Change site title
  tagline: 'Your tagline',           // Change tagline
  url: 'https://docs.yoursite.com',  // Change URL
  // ... more config options
};
```

### Update Navigation

Edit `sidebars.ts` to change sidebar structure.

Edit `docusaurus.config.ts` navbar section to change top navigation.

### Add Custom CSS

Edit `src/css/custom.css`:
```css
:root {
  --ifm-color-primary: #your-brand-color;
}
```

## Common Tasks

### Clear Cache

```bash
npm run clear
```

### Type Check

```bash
npm run typecheck
```

### Format Code

```bash
npm run format
```

### Analyze Bundle

```bash
npm run build:analyze
```

## Configuration Checklist

Before deploying to production:

- [ ] Update `docusaurus.config.ts` with your site info
- [ ] Replace placeholder URLs and emails
- [ ] Add your Algolia search credentials (optional)
- [ ] Add Google Analytics tracking ID (optional)
- [ ] Update footer links and copyright
- [ ] Add your logo to `static/img/`
- [ ] Test all links work
- [ ] Run `npm run build` successfully

## Environment Variables (Optional)

Create `.env` file:
```bash
ALGOLIA_APP_ID=your_app_id
ALGOLIA_API_KEY=your_api_key
GA_TRACKING_ID=G-XXXXXXXXXX
```

## Troubleshooting

### Port Already in Use

```bash
# Kill process on port 3000
lsof -ti:3000 | xargs kill -9

# Or use different port
npm start -- --port 3001
```

### Build Fails

```bash
# Clear cache and reinstall
npm run clear
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Module Not Found

```bash
# Reinstall dependencies
npm ci
```

## Next Steps

1. **Add Content**: Start writing documentation in `docs/` directory
2. **Customize**: Update colors, logos, and branding
3. **Configure Search**: Set up Algolia DocSearch
4. **Deploy**: Deploy to Vercel, Netlify, or GitHub Pages
5. **Monitor**: Set up analytics and monitoring

## Resources

- **Full Documentation**: See `README.md`
- **Deployment Guide**: See `DEPLOYMENT.md`
- **Project Summary**: See `DOCUMENTATION_SUMMARY.md`
- **Docusaurus Docs**: https://docusaurus.io/docs
- **Support**: docs@yourtradingplatform.com

---

Need help? Check the full documentation or contact support.
