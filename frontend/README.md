# MiniSentry Frontend

A modern React frontend for the MiniSentry error monitoring application built with TanStack Router, TanStack Query, and Tailwind CSS.

## Features

- **Modern React 18** with TypeScript and functional components
- **TanStack Router** for type-safe routing with authentication guards
- **TanStack Query** for server state management and caching
- **Tailwind CSS** with custom design system and component library
- **Responsive Design** that works on all devices
- **Authentication System** with JWT token management and auto-refresh
- **Type-Safe API Client** with Axios and comprehensive error handling
- **Hot Module Replacement** for fast development

## Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── ui/           # Reusable UI components (Button, Input, Card, etc.)
│   │   └── layout/       # Layout components (AppLayout, ProtectedRoute)
│   ├── pages/            # Page components (Login, Register, Dashboard)
│   ├── lib/              # Utilities (API client, auth context, utils)
│   ├── types/            # TypeScript type definitions
│   └── styles/           # Global styles and Tailwind configuration
├── public/               # Static assets
└── package.json          # Dependencies and scripts
```

## Technology Stack

### Core
- **React 18** - Modern React with hooks and functional components
- **TypeScript** - Type safety throughout the application
- **Vite** - Fast build tool and development server

### Routing & State
- **TanStack Router** - Type-safe routing with authentication guards
- **TanStack Query** - Server state management and caching

### UI & Styling
- **Tailwind CSS** - Utility-first CSS framework
- **Headless UI** - Unstyled, accessible UI components
- **Heroicons** - Beautiful SVG icons

### API & Forms
- **Axios** - HTTP client with interceptors
- **React Hook Form** - Performant forms with easy validation
- **Zod** - TypeScript-first schema validation

## Getting Started

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Start development server:**
   ```bash
   npm run dev
   ```
   The frontend will be available at `http://localhost:5173`

3. **Build for production:**
   ```bash
   npm run build
   ```

4. **Preview production build:**
   ```bash
   npm run preview
   ```

## Development

### Backend Integration
The frontend is configured to proxy API requests to the backend running on port 8080:

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      secure: false,
    }
  }
}
```

### Authentication Flow
- JWT tokens stored in localStorage
- Automatic token refresh on 401 responses
- Protected routes with authentication guards
- Context-based user state management

### API Client
Type-safe API client with:
- Automatic JWT token injection
- Token refresh handling
- Request/response interceptors
- Comprehensive error handling
- TypeScript types for all endpoints

### Component Library
Reusable UI components built with Tailwind CSS:
- **Button** - Multiple variants (primary, secondary, outline, ghost)
- **Input** - Form inputs with validation states
- **Card** - Content containers with consistent styling
- **Modal** - Accessible modal dialogs
- **Loading** - Loading spinners and states
- **Alert** - Notification components for success/error states

### Design System
Custom Tailwind configuration with:
- Consistent color palette (primary, gray, error, warning, success)
- Typography scale with Inter font
- Spacing and sizing utilities
- Custom shadows and border radius
- Responsive breakpoints

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

## Environment Variables

Create a `.env.local` file for environment-specific configuration:

```env
# API URL (if different from proxy)
VITE_API_URL=http://localhost:8080

# Other environment variables as needed
```

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Contributing

1. Follow the existing code style and patterns
2. Use TypeScript for all new code
3. Write responsive components with Tailwind CSS
4. Ensure proper error handling in API calls
5. Test authentication flows and protected routes