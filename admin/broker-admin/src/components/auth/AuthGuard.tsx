'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { isAuthenticated } from '@/services/apiClient';

interface AuthGuardProps {
    children: React.ReactNode;
}

/**
 * AuthGuard component
 * Protects routes by checking if user is authenticated
 * Redirects to /login if not authenticated
 */
export default function AuthGuard({ children }: AuthGuardProps) {
    const router = useRouter();
    const pathname = usePathname();
    const [isChecking, setIsChecking] = useState(true);
    const [isAuthorized, setIsAuthorized] = useState(false);

    useEffect(() => {
        // Check authentication status
        const checkAuth = () => {
            const authenticated = isAuthenticated();

            if (!authenticated) {
                // Not authenticated - redirect to login
                router.push('/login');
                setIsAuthorized(false);
            } else {
                // Authenticated - allow access
                setIsAuthorized(true);
            }

            setIsChecking(false);
        };

        checkAuth();
    }, [pathname, router]);

    // Show loading state while checking
    if (isChecking) {
        return (
            <div className="min-h-screen bg-[#0A0A0B] flex items-center justify-center">
                <div className="text-center">
                    <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-[#F5C542] to-[#D4AC0D] rounded-lg mb-4">
                        <span className="text-2xl font-bold text-black">RTX</span>
                    </div>
                    <div className="flex items-center justify-center gap-2 text-[#888]">
                        <span className="inline-block w-4 h-4 border-2 border-[#F5C542] border-t-transparent rounded-full animate-spin" />
                        <span className="text-sm">Checking authentication...</span>
                    </div>
                </div>
            </div>
        );
    }

    // Don't render children if not authorized
    if (!isAuthorized) {
        return null;
    }

    // Render protected content
    return <>{children}</>;
}
