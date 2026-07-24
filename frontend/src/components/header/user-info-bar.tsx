"use client";

import { UserDTO } from "@/types";
import { useLanguage } from "@/contexts/theme";

interface UserInfoBarProps {
	partner: UserDTO;
}

export function UserInfoBar({ partner }: UserInfoBarProps) {
	const { t } = useLanguage();
	const isHidden = partner.profile?.is_hidden;

	if (isHidden) {
		return (
			<div className="absolute top-16 left-0 right-0 bg-muted/30 px-4 text-sm text-muted-foreground border-b">
				<div className="flex items-center gap-2">
					<span className="font-medium">{t("anonymous")}</span>
				</div>
			</div>
		);
	}

	return (
		<div className="absolute top-16 left-0 right-0 bg-muted/30 px-4 text-sm text-muted-foreground border-b h-fit">
			<div className="flex items-center gap-2">
				<span className="font-medium">
					{partner.nickname?.trim() || partner.name || t("user")}
				</span>
				{partner.profile?.age && (
					<>
						<span>•</span>
						<span>{t("yearsOld", { age: partner.profile.age })}</span>
					</>
				)}
				{partner.profile?.is_male !== null && partner.profile?.is_male !== undefined && (
					<>
						<span>•</span>
						<span>{partner.profile.is_male ? t("male") : t("female")}</span>
					</>
				)}
			</div>
		</div>
	);
}
