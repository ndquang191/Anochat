"use client";

import * as React from "react";
import {
	Flag,
	Languages,
	LogOut,
	Palette,
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
import { AuroraText } from "@/components/aurora-text";
import { useAuth } from "@/contexts/auth";
import { useAdmin } from "@/contexts/admin";
import { useLanguage, useTheme } from "@/contexts/theme";
import { moderationAPI, userAPI } from "@/lib/api";
import { useUserState, useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { ThemeToggle } from "@/components/theme-toggle";
import { AccountSettingsDialog } from "@/components/account-settings-dialog";
import { LanguageToggle } from "@/components/language-toggle";
import { useAlertDialogContext } from "@/contexts/alert-dialog";
import { toast } from "sonner";

interface UserData {
	id: string;
	email: string;
	name: string;
	age: number | null;
	gender: string;
	isVisible: boolean;
}

const defaultUserData: UserData = {
	id: "",
	email: "",
	name: "...",
	age: null,
	gender: "other",
	isVisible: true,
};

function deriveUserData(
	user: ReturnType<typeof useAuth>["user"],
	data: ReturnType<typeof useUserState>["data"]
): UserData {
	if (!user) return defaultUserData;
	const profile = data?.profile;
	return {
		id: user.id,
		email: user.email || "",
		name: user.name || "Unknown",
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
	const [reported, setReported] = React.useState(false);
	const { setIsAdminOpen } = useAdmin();
	const [localOverrides, setLocalOverrides] = React.useState<Partial<UserData>>({});
	const { t } = useLanguage();
	const { soundEnabled, toggleSound } = useTheme();
	const router = useRouter();
	const { logout, user, room } = useAuth();
	const { data, isLoading } = useUserState();
	const invalidateUserState = useInvalidateUserState();
	const alertDialog = useAlertDialogContext();

	const isAdmin = user?.is_admin === true;
	const partner = room?.partner;
	const derived = deriveUserData(user, data);
	const userData = { ...derived, ...localOverrides };

	React.useEffect(() => {
		setReported(false);
	}, [room?.id]);

	const handleLogout = async () => {
		await logout();
		router.push("/login");
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
		try {
			await userAPI.updateProfile({
				is_hidden: !isVisible,
			});
			setLocalOverrides((prev) => ({ ...prev, isVisible }));
			invalidateUserState();
		} catch {}
	};

	const handleSaveSettings = async (newSettings: {
		age: number | null;
		gender: string;
	}) => {
		await userAPI.updateProfile({
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
				<div
					className="px-6 pb-2 pt-4"
					style={{ fontFamily: "var(--font-changa-one)" }}
				>
					<AuroraText className="text-4xl tracking-widest">ANOCHAT</AuroraText>
				</div>
				<SidebarHeader>
					{state === "expanded" ? (
						<div className="mx-3 my-2 flex flex-col gap-3">
							<div className="rounded-md border border-border/50 bg-card p-4 shadow-sm">
								<div className="flex items-start justify-between gap-3">
									<div className="flex min-w-0 flex-col gap-1">
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
									>
										<Settings size={16} />
									</button>
								</div>
								<div className="mt-3 flex items-center justify-between rounded-md bg-muted/50 px-3 py-2">
									<Label
										htmlFor="public-status"
										className="cursor-pointer select-none text-xs text-muted-foreground"
									>
										{userData.isVisible ? t("publicProfile") : t("privateProfile")}
									</Label>
									<Switch
										id="public-status"
										checked={userData.isVisible}
										onCheckedChange={handleVisibilityToggle}
									/>
								</div>
							</div>
							<div className="flex flex-col gap-2">
								{isAdmin && (
									<button
										type="button"
										onClick={() => setIsAdminOpen(true)}
										className="flex cursor-pointer items-center gap-2 rounded-md border border-border/50 bg-background/60 px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
										title={t("admin")}
									>
										<Shield size={14} />
										<span>{t("admin")}</span>
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
								<div className="mx-3 mb-3 mt-auto rounded-md border border-border/50 bg-card p-4 shadow-sm">
									<div className="flex flex-col items-start gap-3">
										<div className="flex items-center gap-2">
											<Palette size={18} className="text-muted-foreground" />
											<ThemeToggle />
										</div>
										<div className="flex items-center gap-2">
											<Languages size={18} className="text-muted-foreground" />
											<LanguageToggle />
										</div>
										<button
											type="button"
											onClick={toggleSound}
											className="flex cursor-pointer items-center gap-2 text-muted-foreground transition-colors hover:text-foreground"
											title={soundEnabled ? t("turnOffSound") : t("turnOnSound")}
										>
											{soundEnabled ? (
												<Volume2 size={18} />
											) : (
												<VolumeX size={18} />
											)}
											<span className="text-sm">
												{soundEnabled ? t("turnOffSound") : t("turnOnSound")}
											</span>
										</button>
									</div>
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
					age: userData.age,
					gender: userData.gender,
				}}
				onSave={handleSaveSettings}
			/>
		</>
	);
}
