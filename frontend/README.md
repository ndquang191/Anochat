# Anochat - Anonymous Chat Application

A modern anonymous chat application built with Next.js, React, and TypeScript.

## Features

- **Modern UI**: Built with shadcn/ui components and Tailwind CSS
- **Anonymous Chat**: Connect with random users for anonymous conversations
- **User Settings**: Customizable user profile with privacy controls
- **Responsive Design**: Works on desktop and mobile devices
- **Real-time Messaging**: Socket.io integration for live chat functionality

## Tech Stack

- **Frontend**: Next.js 15, React 19, TypeScript
- **UI Components**: shadcn/ui, Radix UI, Tailwind CSS
- **Real-time**: Socket.io Client
- **State Management**: React Hooks
- **Styling**: Tailwind CSS with custom animations

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd anochat
```

2. Install dependencies:
```bash
npm install
```

3. Set up environment variables:
Create a `.env.local` file in the root directory:
```env
NEXT_PUBLIC_API_URL=http://localhost:8000
```

4. Run the development server:
```bash
npm run dev
```

5. Open [http://localhost:3000](http://localhost:3000) in your browser.

## Project Structure

```
src/
├── app/                    # Next.js app directory
│   ├── (auth)/            # Authentication pages
│   ├── (main)/            # Main application pages
│   └── globals.css        # Global styles
├── components/            # React components
│   ├── ui/               # shadcn/ui components
│   ├── app-sidebar.tsx   # Main sidebar component
│   ├── chat-box.tsx      # Chat interface
│   ├── header.tsx        # Application header
│   ├── login-form.tsx    # Login form
│   └── user-settings-dialog.tsx # User settings dialog
├── hooks/                # Custom React hooks
├── lib/                  # Utility functions
└── middleware.ts         # Next.js middleware
```

## Development

### Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run lint` - Run ESLint

### Architecture Notes

- **Authentication**: Currently using placeholder authentication system. Ready for integration with your preferred auth provider.
- **Real-time Chat**: Uses Socket.io for real-time messaging capabilities.
- **User Management**: Mock user data system in place, ready for backend integration.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.
