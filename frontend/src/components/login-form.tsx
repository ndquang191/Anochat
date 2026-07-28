"use client";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { ShineBorder } from "@/components/ui/shine-border";
import { useEffect, useRef, useState } from "react";
import { authAPI } from "@/lib/api";
import { BrandLogo } from "@/components/brand-logo";
import { useAuth } from "@/contexts/auth";
import { useRouter } from "next/navigation";
import { useLanguage } from "@/contexts/theme";
import { toast } from "sonner";

function GoogleIcon() {
	return (
		<svg viewBox="0 0 24 24" className="w-4 h-4" fill="currentColor" aria-hidden="true">
			<path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
			<path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
			<path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" />
			<path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
		</svg>
	);
}

export function LoginForm({ className, ...props }: React.ComponentPropsWithoutRef<"div">) {
	const [isLoading, setIsLoading] = useState(false);
	const apiAvailableRef = useRef(false);
	const apiHealthCheckRef = useRef<Promise<void> | null>(null);
	const { login } = useAuth();
	const { t } = useLanguage();
	const router = useRouter();
	const devLoginEnabled = process.env.NEXT_PUBLIC_DEV_AUTH_ENABLED === "true";

	useEffect(() => {
		const healthCheck = authAPI.assertAvailable();
		apiHealthCheckRef.current = healthCheck;

		void healthCheck
			.then(() => {
				apiAvailableRef.current = true;
			})
			.catch(() => {
				apiAvailableRef.current = false;
			})
			.finally(() => {
				if (apiHealthCheckRef.current === healthCheck) {
					apiHealthCheckRef.current = null;
				}
			});
	}, []);

	const handleGoogleLogin = async (e: React.FormEvent) => {
		e.preventDefault();
		setIsLoading(true);
		try {
			if (!apiAvailableRef.current) {
				await (apiHealthCheckRef.current ?? authAPI.assertAvailable());
				apiAvailableRef.current = true;
			}
			window.location.assign(authAPI.getGoogleAuthUrl());
		} catch (error) {
			toast.error(error instanceof Error ? error.message : t("apiUnavailable"));
			setIsLoading(false);
		}
	};

	const handleDevLogin = async (user: "a" | "b") => {
		setIsLoading(true);
		try {
			await authAPI.devLogin(user);
			await login();
			router.replace("/");
		} catch (error) {
			toast.error(error instanceof Error ? error.message : t("apiUnavailable"));
			setIsLoading(false);
		}
	};

	return (
		<div className={cn("flex flex-col gap-6", className)} {...props}>
			<Card className="relative gap-5 overflow-hidden py-7">
				<ShineBorder shineColor={["hsl(var(--primary))", "hsl(var(--aurora-1))", "hsl(var(--aurora-3))"]} duration={10} />
				<CardHeader className="flex flex-col items-center gap-2 px-8 py-2 text-center">
					<div className="flex items-center justify-center gap-3">
						<BrandLogo showSlogan={false} iconClassName="size-12" />
						<h1 className="text-4xl leading-none tracking-tight text-primary [font-family:var(--font-changa-one)]">
							AnoChat
						</h1>
					</div>
					<p className="text-sm leading-relaxed text-muted-foreground">
						{t("signInDescription")}
					</p>
				</CardHeader>
				<CardContent className="px-8">
					<form onSubmit={handleGoogleLogin} className="space-y-3">
						<Button type="submit" className="h-11 w-full cursor-pointer" disabled={isLoading}>
							{isLoading ? (
								t("signingIn")
							) : (
								<>
									<GoogleIcon />
									<span>{t("signInGoogle")}</span>
								</>
							)}
						</Button>
						{devLoginEnabled && (
							<div className="grid grid-cols-2 gap-2">
								<Button type="button" variant="outline" disabled={isLoading} onClick={() => handleDevLogin("a")}>
									Dev A
								</Button>
								<Button type="button" variant="outline" disabled={isLoading} onClick={() => handleDevLogin("b")}>
									Dev B
								</Button>
							</div>
						)}
					</form>
				</CardContent>
			</Card>
		</div>
	);
}
