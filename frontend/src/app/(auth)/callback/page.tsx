"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/contexts/auth";
import { getCookie, deleteCookie } from "@/lib/cookies";

export default function CallbackPage() {
	const router = useRouter();
	const { login } = useAuth();
	const [error, setError] = useState<string | null>(null);
	const [retryCount, setRetryCount] = useState(0);

	useEffect(() => {
		const handleCallback = async () => {
			try {
				// Get user data from temporary cookie set by backend
				const userDataStr = getCookie("temp_user_data");

				if (userDataStr) {
					try {
						// Decode URL encoded cookie
						const decodedUserDataStr = decodeURIComponent(userDataStr);

						const userData = JSON.parse(decodedUserDataStr);

						// Token is already set in HTTP-only cookie by backend
						// Just update auth context with user data
						login("", userData); // Empty token since it's in cookie

						// Clear temporary cookie
						deleteCookie("temp_user_data");

						// Add small delay to ensure state is updated
						setTimeout(() => {
							router.push("/");
						}, 100);
					} catch (parseError) {
						console.error("DEBUG: Parse error:", parseError);
						console.error("DEBUG: Raw userDataStr:", userDataStr);
						setError("Invalid user data received");
					}
				} else {
					// Retry mechanism - wait a bit and try again
					if (retryCount < 3) {
						setTimeout(() => {
							setRetryCount((prev) => prev + 1);
						}, 1000);
					} else {
						setError("No user data received from authentication. Please try logging in again.");
					}
				}
			} catch (error) {
				console.error("DEBUG: Error stack:", error instanceof Error ? error.stack : "No stack trace");
				setError("Authentication failed");
			}
		};

		// Run when component mounts or retry count changes
		handleCallback();
	}, [retryCount, login, router]); // Dependencies for effect


	if (error) {
		return (
			<div className="flex min-h-svh w-full items-center justify-center p-6">
				<div className="text-center">
					<h1 className="text-2xl font-bold text-red-600 mb-4">Authentication Error</h1>
					<p className="text-gray-600 mb-4">{error}</p>
					<button
						onClick={() => router.push("/login")}
						className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
					>
						Try Again
					</button>
				</div>
			</div>
		);
	}

	// Show loading while retrying
	if (retryCount > 0) {
		return (
			<div className="flex min-h-svh w-full items-center justify-center p-6">
				<div className="text-center">
					<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
					<h1 className="text-2xl font-bold mb-2">Processing...</h1>
					<p className="text-gray-600">Retry {retryCount}/3</p>
				</div>
			</div>
		);
	}

	// Return null to let loading.tsx handle the loading state
	return null;
}
