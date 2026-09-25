"use client";

import { useState } from "react";

import {
	ChevronUp,
	Languages,
	Palette,
	SlidersHorizontal,
	Volume2,
	VolumeX,
	Download,
	Bell,
} from "lucide-react";
import {
	LanguageToggle,
	changeLanguageWithTransition,
} from "@/components/language-toggle";
import {
	ThemeToggle,
	changeThemeWithTransition,
} from "@/components/theme-toggle";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useLanguage, useTheme } from "@/contexts/theme";
import { cn } from "@/lib/utils";
import { userAPI } from "@/lib/api";
import { useUserState, useInvalidateUserState } from "@/hooks/queries/use-user-state";
import type { MatchMode } from "@/types";
import { toast } from "sonner";
import { usePWA } from "@/contexts/pwa";
import { useAlertDialogContext } from "@/contexts/alert-dialog";

const themeOrder = ["blue", "dark", "pink"] as const;
const iconButtonClass =
	"shrink-0 cursor-pointer rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring";

interface SidebarPreferencesProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	panelId: string;
}

export function SidebarPreferences({
	open,
	onOpenChange,
	panelId,
}: SidebarPreferencesProps) {
	const { language, setLanguage, t } = useLanguage();
	const { theme, setTheme, soundEnabled, toggleSound } = useTheme();
	const { data } = useUserState();
	const invalidateUserState = useInvalidateUserState();
	const [isUpdatingMatchMode, setIsUpdatingMatchMode] = useState(false);
	const { canInstall, isIOS, isStandalone, install, pushEnabled, isPushSubscribed, ensurePushSubscription, unsubscribe } = usePWA();
	const alertDialog = useAlertDialogContext();
	const soundLabel = soundEnabled ? t("turnOffSound") : t("turnOnSound");

	const cycleTheme = (origin: HTMLElement) => {
		const currentIndex = themeOrder.indexOf(theme);
		const nextTheme = themeOrder[(currentIndex + 1) % themeOrder.length];
		changeThemeWithTransition(nextTheme, theme, setTheme, origin);
	};

	const cycleLanguage = (origin: HTMLElement) => {
		const nextLanguage = language === "vi" ? "en" : "vi";
		changeLanguageWithTransition(nextLanguage, language, setLanguage, origin);
	};

	const updateMatchMode = async (mode: MatchMode) => {
		setIsUpdatingMatchMode(true);
		try {
			await userAPI.updateMatchPreference(mode);
			invalidateUserState();
			toast.success(t("matchPreferenceSaved"));
		} catch {
			toast.error(t("matchPreferenceSaveFailed"));
		} finally {
			setIsUpdatingMatchMode(false);
		}
	};

	return (
		<div className="mx-3 mb-3 mt-auto">
			<div
				className={cn(
					"grid transition-[grid-template-rows,opacity,margin] duration-200 ease-out",
					open
						? "mb-2 grid-rows-[1fr] opacity-100"
						: "mb-0 grid-rows-[0fr] opacity-0",
				)}
			>
				<div className="overflow-hidden">
					<div
						id={panelId}
						aria-hidden={!open}
						inert={!open}
						className="rounded-md border border-border/50 bg-card p-4 shadow-sm"
					>
						<div className="flex flex-col gap-3">
							{canInstall && (
								<button
									type="button"
									className="flex min-h-8 items-center gap-3 text-left text-sm"
									onClick={async () => {
										if (isIOS) {
											await alertDialog.open({ title: language === "vi" ? "Cài đặt AnoChat" : "Install AnoChat", description: language === "vi" ? "Mở menu Chia sẻ của trình duyệt rồi chọn “Thêm vào Màn hình chính”." : "Open the browser Share menu and choose “Add to Home Screen”.", confirmText: t("confirm") });
											return;
										}
										await install();
									}}
								>
									<Download size={16} aria-hidden="true" />
									<span>{language === "vi" ? "Cài đặt AnoChat" : "Install AnoChat"}</span>
								</button>
							)}
							{pushEnabled && (
								<div className="flex min-h-8 items-center gap-3">
									<Bell size={16} aria-hidden="true" />
									<Switch
										checked={isPushSubscribed}
										onCheckedChange={async (checked) => {
											if (!checked) { await unsubscribe(); return; }
											if (isIOS && !isStandalone) {
												await alertDialog.open({ title: language === "vi" ? "Cài đặt AnoChat" : "Install AnoChat", description: language === "vi" ? "Trên iPhone/iPad, hãy thêm AnoChat vào Màn hình chính trước khi bật thông báo." : "On iPhone/iPad, add AnoChat to the Home Screen before enabling notifications.", confirmText: t("confirm") });
												return;
											}
											await ensurePushSubscription();
										}}
										aria-label={language === "vi" ? "Thông báo ghép đôi" : "Match notifications"}
									/>
									<span className="text-sm">{language === "vi" ? "Thông báo ghép đôi" : "Match notifications"}</span>
								</div>
							)}
							<div className="flex items-center gap-3">
								<button
									type="button"
									onClick={(event) => cycleTheme(event.currentTarget)}
									className={iconButtonClass}
									aria-label={t("theme")}
									title={t("theme")}
								>
									<Palette size={16} aria-hidden="true" />
								</button>
								<ThemeToggle />
							</div>
							<div className="flex items-center gap-3">
								<button
									type="button"
									onClick={(event) => cycleLanguage(event.currentTarget)}
									className={iconButtonClass}
									aria-label={t("interfaceLanguage")}
									title={t("interfaceLanguage")}
								>
									<Languages size={16} aria-hidden="true" />
								</button>
								<LanguageToggle />
							</div>
							<div className="flex min-h-8 items-center gap-3">
								<button
									type="button"
									onClick={toggleSound}
									className={iconButtonClass}
									aria-label={soundLabel}
									title={soundLabel}
								>
									{soundEnabled ? (
										<Volume2 size={16} aria-hidden="true" />
									) : (
										<VolumeX size={16} aria-hidden="true" />
									)}
								</button>
								<Switch
									checked={soundEnabled}
									onCheckedChange={toggleSound}
									aria-label={soundLabel}
									title={soundLabel}
								/>
							</div>
							{data?.match_settings?.allow_user_choice && (
								<div className="space-y-2 border-t pt-3">
									<label className="text-xs font-medium text-muted-foreground">{t("matchMode")}</label>
									<Select
										value={data.match_settings.user_preference ?? data.match_settings.default_mode}
										onValueChange={(value) => void updateMatchMode(value as MatchMode)}
										disabled={isUpdatingMatchMode}
									>
										<SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
										<SelectContent>
											<SelectItem value="mixed">{t("matchModeMixed")}</SelectItem>
											<SelectItem value="opposite_sex">{t("matchModeOpposite")}</SelectItem>
										</SelectContent>
									</Select>
								</div>
							)}
						</div>
					</div>
				</div>
			</div>
			<button
				type="button"
				onClick={() => onOpenChange(!open)}
				aria-expanded={open}
				aria-controls={panelId}
				className="flex w-full cursor-pointer items-center justify-between rounded-md border border-border/50 bg-background/60 px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
			>
				<span className="flex items-center gap-2">
					<SlidersHorizontal size={16} />
					<span>{t("settings")}</span>
				</span>
				<ChevronUp
					size={16}
					className={cn(
						"transition-transform duration-200",
						open && "rotate-180",
					)}
				/>
			</button>
		</div>
	);
}
