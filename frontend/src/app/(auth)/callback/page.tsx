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
				const userDataStr = getCookie("temp_user_data");

				if (userDataStr) {
					try {
						const decodedUserDataStr = decodeURIComponent(userDataStr);
						const userData = JSON.parse(decodedUserDataStr);

						login(userData);
						deleteCookie("temp_user_data");

						setTimeout(() => {
							router.push("/");
						}, 100);
					} catch {
						setError("Invalid user data received");
					}
				} else {
					if (retryCount < 3) {
						setTimeout(() => {
							setRetryCount((prev) => prev + 1);
						}, 1000);
					} else {
						setError("No user data received from authentication. Please try logging in again.");
					}
				}
			} catch {
				setError("Authentication failed");
			}
		};

		handleCallback();
	}, [retryCount, login, router]);


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

	return null;
}
