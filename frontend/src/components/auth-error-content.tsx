"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useLanguage } from "@/contexts/theme";

export function AuthErrorContent() {
	const { t } = useLanguage();

	return (
		<div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
			<div className="w-full max-w-sm">
				<Card>
					<CardHeader>
						<CardTitle className="text-2xl">{t("somethingWentWrong")}</CardTitle>
					</CardHeader>
					<CardContent className="flex flex-col gap-4">
						<p className="text-sm text-muted-foreground">
							{t("comeBackLaterOrSignOut")}
						</p>
						<Button asChild>
							<Link href="/login">{t("signOut")}</Link>
						</Button>
					</CardContent>
				</Card>
			</div>
		</div>
	);
}
