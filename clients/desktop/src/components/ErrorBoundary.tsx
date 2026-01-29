import { Component } from 'react';
import type { ErrorInfo, ReactNode } from 'react';

interface Props {
    children?: ReactNode;
    fallback?: ReactNode;
}

interface State {
    hasError: boolean;
    error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
    public state: State = {
        hasError: false,
        error: null
    };

    public static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error("Uncaught error:", error, errorInfo);
    }

    public render() {
        if (this.state.hasError) {
            if (this.props.fallback) {
                return this.props.fallback;
            }
            return (
                <div className="flex flex-col items-center justify-center p-4 bg-zinc-900 text-red-500 rounded border border-red-500/20">
                    <h2 className="text-lg font-bold mb-2">Something went wrong</h2>
                    <pre className="text-xs bg-black/50 p-2 rounded max-w-full overflow-auto">
                        {this.state.error?.message}
                    </pre>
                </div>
            );
        }

        return this.props.children;
    }
}
