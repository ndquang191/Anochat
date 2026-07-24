"use client";

import * as React from "react";
import {
	ChevronUp,
	Flag,
	Languages,
	LogOut,
	MessageCircle,
	Palette,
	SlidersHorizontal,
	Volume2,
	VolumeX,
	Settings,
	Shield,
	User,
} from "lucide-react";
import {
	Sidebar,
	SidebarContent,
	SidebarFooter,
	SidebarGroup,
	SidebarGroupContent,
	SidebarHeader,
	SidebarRail,
	useSidebar,
} from "@/components/ui/sidebar";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { BrandLogo } from "@/components/brand-logo";
import { useAuth } from "@/contexts/auth";
import { useAdmin } from "@/contexts/admin";
import { useLanguage, useTheme } from "@/contexts/theme";
import { moderationAPI, userAPI } from "@/lib/api";
import { useUserState, useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { changeThemeWithTransition, ThemeToggle } from "@/components/theme-toggle";
import { AccountSettingsDialog } from "@/components/account-settings-dialog";
import {
	changeLanguageWithTransition,
	LanguageToggle,
} from "@/components/language-toggle";
import { useAlertDialogContext } from "@/contexts/alert-dialog";
import { toast } from "sonner";

interface UserData {
	id: string;
	email: string;
	name: string;
	nickname: string;
	nicknameChangeAvailableAt: number | null;
	age: number | null;
	gender: string;
	isVisible: boolean;
}

const defaultUserData: UserData = {
	id: "",
	email: "",
	name: "...",
	nickname: "",
	nicknameChangeAvailableAt: null,
	age: null,
	gender: "other",
	isVisible: true,
};

const themeOrder = ["blue", "dark", "pink"] as const;
const preferenceIconButtonClass =
	"shrink-0 rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring cursor-pointer";

function deriveUserData(
	user: ReturnType<typeof useAuth>["user"],
	data: ReturnType<typeof useUserState>["data"],
	fallbackName: string
): UserData {
	if (!user) return defaultUserData;
	const profile = data?.profile;
	const nickname = profile?.nickname?.trim() ?? "";
	return {
		id: user.id,
		email: user.email || "",
		name: nickname || user.name || fallbackName,
		nickname,
		nicknameChangeAvailableAt: profile?.nickname_change_available_at ?? null,
		age: profile?.age ?? null,
		gender:
			profile?.is_male === true
				? "male"
				: profile?.is_male === false
					? "female"
					: "other",
		isVisible: !(profile?.is_hidden ?? false),
	};
}

export function AppShellSidebar({
	className,
	...props
}: React.HTMLAttributes<HTMLDivElement>) {
	const { state } = useSidebar();
	const [isSettingsOpen, setIsSettingsOpen] = React.useState(false);
	const [isPreferencesOpen, setIsPreferencesOpen] = React.useState(false);
	const [isVisibilityUpdating, setIsVisibilityUpdating] = React.useState(false);
	const preferencesPanelId = React.useId();
	const [reported, setReported] = React.useState(false);
	const { isAdminOpen, setIsAdminOpen } = useAdmin();
	const [localOverrides, setLocalOverrides] = React.useState<Partial<UserData>>({});
	const { language, setLanguage, t } = useLanguage();
	const { theme, setTheme, soundEnabled, toggleSound } = useTheme();
	const router = useRouter();
	const { logout, user, room } = useAuth();
	const { data, isLoading } = useUserState();
	const invalidateUserState = useInvalidateUserState();
	const alertDialog = useAlertDialogContext();

	const isAdmin = user?.is_admin === true;
	const partner = room?.partner;
	const derived = deriveUserData(user, data, t("user"));
	const userData = { ...derived, ...localOverrides };

	React.useEffect(() => {
		setReported(false);
	}, [room?.id]);

	const handleLogout = async () => {
		await logout();
		router.push("/login");
	};

	const cycleTheme = (origin: HTMLElement) => {
		const currentIndex = themeOrder.indexOf(theme);
		const nextTheme = themeOrder[(currentIndex + 1) % themeOrder.length];
		changeThemeWithTransition(nextTheme, theme, setTheme, origin);
	};

	const cycleLanguage = (origin: HTMLElement) => {
		const nextLanguage = language === "vi" ? "en" : "vi";
		changeLanguageWithTransition(nextLanguage, language, setLanguage, origin);
	};

	const getGenderDisplay = (genderValue: string) => {
		switch (genderValue) {
			case "male":
				return t("male");
			case "female":
				return t("female");
			default:
				return t("other");
		}
	};

	const handleVisibilityToggle = async (isVisible: boolean) => {
		if (isVisibilityUpdating) return;

		const previousVisibility = userData.isVisible;
		setLocalOverrides((prev) => ({ ...prev, isVisible }));
		setIsVisibilityUpdating(true);

		try {
			await userAPI.updateProfile({
				is_hidden: !isVisible,
			});
			invalidateUserState();
		} catch {
			setLocalOverrides((prev) => ({
				...prev,
				isVisible: previousVisibility,
			}));
			toast.error(t("somethingWentWrong"));
		} finally {
			setIsVisibilityUpdating(false);
		}
	};

	const handleSaveSettings = async (newSettings: {
		nickname: string;
		age: number | null;
		gender: string;
	}) => {
		const response = await userAPI.updateProfile({
			nickname: newSettings.nickname,
			age: newSettings.age,
			is_male:
				newSettings.gender === "male"
					? true
					: newSettings.gender === "female"
						? false
						: undefined,
		});

		setLocalOverrides((prev) => ({
			...prev,
			name: newSettings.nickname || user?.name || t("user"),
			nickname: newSettings.nickname,
			nicknameChangeAvailableAt:
				response.data?.nickname_change_available_at ??
				prev.nicknameChangeAvailableAt ??
				null,
			age: newSettings.age,
			gender: newSettings.gender,
		}));
		invalidateUserState();
		setIsSettingsOpen(false);
	};

	const handleReport = async () => {
		if (!partner || !room || reported) return;

		const confirmed = await alertDialog.open({
			title: t("reportConfirmTitle"),
			description: t("reportConfirmDescription"),
			confirmText: t("report"),
			cancelText: t("cancel"),
		});

		if (!confirmed) return;

		try {
			await moderationAPI.createReport(partner.id, room.id);
			setReported(true);
			toast.success(t("reportSubmitted"));
		} catch {}
	};

	return (
		<>
			<Sidebar className={cn(className)} {...props}>
				<BrandLogo
					className={cn(
						"pb-2 pt-4",
						state === "collapsed" ? "justify-center px-2" : "px-6"
					)}
					showSlogan={state === "expanded"}
				/>
				<SidebarHeader>
					{state === "expanded" ? (
						<div className="mx-3 my-2 flex flex-col gap-3">
							<div className="rounded-md border border-border/50 bg-card p-4 shadow-sm">
								<div
									aria-hidden={!userData.isVisible}
									className={cn(
										"grid transition-[grid-template-rows,opacity] duration-300 ease-out",
										userData.isVisible
											? "grid-rows-[1fr] opacity-100"
											: "grid-rows-[0fr] opacity-0"
									)}
								>
									<div className="overflow-hidden">
										<div
											className={cn(
												"flex items-start justify-between gap-3 transition-transform duration-300 ease-out",
												userData.isVisible
													? "translate-y-0"
													: "-translate-y-2"
											)}
										>
											<div
												className="flex min-w-0 flex-1 flex-col gap-1"
											>
												{isLoading && !data ? (
													<Skeleton className="h-5 w-28 rounded" />
												) : (
													<span className="truncate text-base font-semibold leading-tight">
														{userData.name}
													</span>
												)}
												<div className="flex items-center gap-1.5 text-muted-foreground">
													<User size={12} />
													{isLoading && !data ? (
														<Skeleton className="h-3.5 w-16 rounded" />
													) : (
														<span className="text-xs">
															{getGenderDisplay(userData.gender)}
															{userData.age
																? `, ${t("yearsOld", { age: userData.age })}`
																: ""}
														</span>
													)}
												</div>
											</div>
											<button
												type="button"
												onClick={() => setIsSettingsOpen(true)}
												className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border/60 bg-background/70 text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
												aria-label={t("accountSettings")}
												title={t("accountSettings")}
												tabIndex={userData.isVisible ? 0 : -1}
											>
												<Settings size={16} />
											</button>
										</div>
									</div>
								</div>
								<div
									className={cn(
										"flex items-center justify-between rounded-md bg-muted/50 px-3 py-2 transition-[margin] duration-300 ease-out",
										userData.isVisible ? "mt-3" : "mt-0"
									)}
								>
									<Label
										htmlFor="private-status"
										className="cursor-pointer select-none text-xs text-muted-foreground"
									>
										{t("privateProfile")}
									</Label>
									<Switch
										id="private-status"
										checked={!userData.isVisible}
										onCheckedChange={(isPrivate) =>
											handleVisibilityToggle(!isPrivate)
										}
										disabled={isVisibilityUpdating}
									/>
								</div>
							</div>
							<div className="flex flex-col gap-2">
								{isAdmin && (
									<button
										type="button"
										onClick={() => setIsAdminOpen(!isAdminOpen)}
										className="flex cursor-pointer items-center gap-2 rounded-md border border-border/50 bg-background/60 px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
										title={isAdminOpen ? t("chat") : t("admin")}
									>
										{isAdminOpen ? <MessageCircle size={14} /> : <Shield size={14} />}
										<span>{isAdminOpen ? t("chat") : t("admin")}</span>
									</button>
								)}
							{partner && !reported && (
								<button
									type="button"
									onClick={handleReport}
									className="flex cursor-pointer items-center gap-2 rounded-md border border-border/50 bg-background/60 px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-950"
									title={t("reportUser")}
								>
									<Flag size={14} />
									<span>{t("reportUser")}</span>
								</button>
							)}
								<button
									type="button"
									onClick={handleLogout}
									className="flex cursor-pointer items-center gap-2 rounded-md border border-border/50 bg-background/60 px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
									title={t("logout")}
								>
									<LogOut size={14} />
									<span>{t("logout")}</span>
								</button>
							</div>
						</div>
					) : (
						<div className="flex justify-center py-3">
							<User size={18} className="text-muted-foreground" />
						</div>
					)}
				</SidebarHeader>
				<SidebarContent>
					{state === "expanded" && (
						<SidebarGroup className="mt-auto">
							<SidebarGroupContent>
								<div className="mx-3 mb-3 mt-auto">
									<div
										className={cn(
											"grid transition-[grid-template-rows,opacity,margin] duration-200 ease-out",
											isPreferencesOpen
												? "mb-2 grid-rows-[1fr] opacity-100"
												: "mb-0 grid-rows-[0fr] opacity-0"
										)}
									>
										<div className="overflow-hidden">
											<div
												id={preferencesPanelId}
												aria-hidden={!isPreferencesOpen}
												inert={!isPreferencesOpen}
												className="rounded-md border border-border/50 bg-card p-4 shadow-sm"
											>
												<div className="flex flex-col gap-3">
													<div className="flex items-center justify-start gap-3">
														<button
															type="button"
															onClick={(event) =>
																cycleTheme(event.currentTarget)
															}
															className={preferenceIconButtonClass}
															aria-label={t("theme")}
															title={t("theme")}
														>
															<Palette size={16} aria-hidden="true" />
														</button>
														<ThemeToggle />
													</div>
													<div className="flex items-center justify-start gap-3">
														<button
															type="button"
															onClick={(event) =>
																cycleLanguage(event.currentTarget)
															}
															className={preferenceIconButtonClass}
															aria-label={t("interfaceLanguage")}
															title={t("interfaceLanguage")}
														>
															<Languages size={16} aria-hidden="true" />
														</button>
														<LanguageToggle />
													</div>
													<div className="flex min-h-8 items-center justify-start gap-3">
														<button
															type="button"
															onClick={toggleSound}
															className={preferenceIconButtonClass}
															aria-label={
																soundEnabled
																	? t("turnOffSound")
																	: t("turnOnSound")
															}
															title={
																soundEnabled
																	? t("turnOffSound")
																	: t("turnOnSound")
															}
														>
															{soundEnabled ? (
																<Volume2 size={16} aria-hidden="true" />
															) : (
																<VolumeX size={16} aria-hidden="true" />
															)}
														</button>
														<Switch
															checked={soundEnabled}
															onCheckedChange={() => toggleSound()}
															aria-label={
																soundEnabled
																	? t("turnOffSound")
																	: t("turnOnSound")
															}
															title={
																soundEnabled
																	? t("turnOffSound")
																	: t("turnOnSound")
															}
														/>
													</div>
												</div>
											</div>
										</div>
									</div>
									<button
										type="button"
										onClick={() => setIsPreferencesOpen((open) => !open)}
										aria-expanded={isPreferencesOpen}
										aria-controls={preferencesPanelId}
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
												isPreferencesOpen && "rotate-180"
											)}
										/>
									</button>
								</div>
							</SidebarGroupContent>
						</SidebarGroup>
					)}
				</SidebarContent>
				<SidebarFooter />
				<SidebarRail />
			</Sidebar>

			<AccountSettingsDialog
				open={isSettingsOpen}
				onOpenChange={setIsSettingsOpen}
				initialData={{
					nickname: userData.nickname,
					nicknameChangeAvailableAt: userData.nicknameChangeAvailableAt,
					age: userData.age,
					gender: userData.gender,
				}}
				onSave={handleSaveSettings}
			/>
		</>
	);
}
