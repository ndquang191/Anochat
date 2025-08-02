"use client";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useState } from "react";

export function LoginForm({ className, ...props }: React.ComponentPropsWithoutRef<"div">) {
	const [isLoading, setIsLoading] = useState(false);

	const handleSocialLogin = async (e: React.FormEvent) => {
		e.preventDefault();
		setIsLoading(true);

		try {
			// Redirect to backend OAuth endpoint
			window.location.href = "http://localhost:8080/auth/google";
		} catch (error) {
			console.error("Login error:", error);
			setIsLoading(false);
		}
	};

	return (
		<div className={cn("flex flex-col gap-6", className)} {...props}>
			<Card>
				<CardHeader>
					<CardTitle className="text-2xl">Welcome!</CardTitle>
					<CardDescription>Sign in to your account to continue</CardDescription>
				</CardHeader>
				<CardContent>
					<form onSubmit={handleSocialLogin}>
						<div className="flex flex-col gap-6">
							<Button type="submit" className="w-full" disabled={isLoading}>
								{isLoading ? "Logging in..." : "Continue with Google"}
							</Button>
						</div>
					</form>
				</CardContent>
			</Card>
		</div>
	);
}
