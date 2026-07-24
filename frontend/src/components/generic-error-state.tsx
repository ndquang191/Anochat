"use client";

import { AlertTriangle, LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/auth";
import { useLanguage } from "@/contexts/theme";

export function GenericErrorState() {
	const { logout } = useAuth();
	const { t } = useLanguage();

	return (
		<div className="flex h-full w-full items-center justify-center p-6">
			<div className="flex max-w-sm flex-col items-center gap-3 text-center">
				<AlertTriangle className="size-6 text-muted-foreground" aria-hidden="true" />
				<h2 className="text-base font-semibold">{t("somethingWentWrong")}</h2>
				<p className="text-sm text-muted-foreground">
					{t("comeBackLaterOrSignOut")}
				</p>
				<Button type="button" variant="outline" onClick={() => void logout()}>
					<LogOut className="size-4" />
					{t("logout")}
				</Button>
			</div>
		</div>
	);
}
