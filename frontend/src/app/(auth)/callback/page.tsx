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
				console.log("DEBUG: Callback page loaded, retry count:", retryCount);
				console.log("DEBUG: All cookies:", document.cookie);
				console.log("DEBUG: Current URL:", window.location.href);

				// Get user data from temporary cookie set by backend
				const userDataStr = getCookie("temp_user_data");
				console.log("DEBUG: temp_user_data cookie:", userDataStr);

				if (userDataStr) {
					try {
						// Decode URL encoded cookie
						const decodedUserDataStr = decodeURIComponent(userDataStr);
						console.log("DEBUG: Decoded user data string:", decodedUserDataStr);

						const userData = JSON.parse(decodedUserDataStr);
						console.log("DEBUG: Parsed user data:", userData);

						// Token is already set in HTTP-only cookie by backend
						// Just update auth context with user data
						login("", userData); // Empty token since it's in cookie

						// Clear temporary cookie
						deleteCookie("temp_user_data");
						console.log("DEBUG: Cleared temp_user_data cookie");

						// Add small delay to ensure state is updated
						setTimeout(() => {
							console.log("DEBUG: Redirecting to /");
							router.push("/");
						}, 100);
					} catch (parseError) {
						console.error("DEBUG: Parse error:", parseError);
						console.error("DEBUG: Raw userDataStr:", userDataStr);
						setError("Invalid user data received");
					}
				} else {
					console.log("DEBUG: No temp_user_data cookie found");
					console.log(
						"DEBUG: Available cookies:",
						document.cookie.split(";").map((c) => c.trim())
					);

					// Retry mechanism - wait a bit and try again
					if (retryCount < 3) {
						console.log("DEBUG: Retrying in 1 second...");
						setTimeout(() => {
							setRetryCount((prev) => prev + 1);
						}, 1000);
					} else {
						console.error("DEBUG: Max retries reached, showing error");
						setError("No user data received from authentication. Please try logging in again.");
					}
				}
			} catch (error) {
				console.error("DEBUG: Callback error:", error);
				console.error("DEBUG: Error stack:", error instanceof Error ? error.stack : "No stack trace");
				setError("Authentication failed");
			}
		};

		// Run when component mounts or retry count changes
		handleCallback();
	}, [retryCount]); // Add retryCount as dependency

	// Log component render
	console.log("DEBUG: CallbackPage render, error:", error, "retryCount:", retryCount);

	if (error) {
		console.log("DEBUG: Rendering error UI");
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
		console.log("DEBUG: Rendering retry UI");
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
	console.log("DEBUG: Rendering null (loading)");
	return null;
}
