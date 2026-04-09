"use client";

import * as React from "react";
import {
	Languages,
	LogOut,
	Palette,
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
	SidebarGroupLabel,
	SidebarHeader,
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
	SidebarRail,
	useSidebar,
} from "@/components/ui/sidebar";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { AdminUserID } from "@/types";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { AuroraText } from "@/components/aurora-text";
import { useAuth } from "@/contexts/auth";
import { useAdmin } from "@/contexts/admin";
import { useLanguage } from "@/contexts/theme";
import { userAPI } from "@/lib/api";
import { useUserState, useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { ThemeToggle } from "@/components/theme-toggle";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { AccountSettingsDialog } from "@/components/account-settings-dialog";

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
	const { setIsAdminOpen } = useAdmin();
	const [localOverrides, setLocalOverrides] = React.useState<Partial<UserData>>({});
	const { t, language, setLanguage } = useLanguage();
	const router = useRouter();
	const { logout, user } = useAuth();
	const { data, isLoading } = useUserState();
	const invalidateUserState = useInvalidateUserState();

	const isAdmin = user?.id === AdminUserID;
	const derived = deriveUserData(user, data);
	const userData = { ...derived, ...localOverrides };

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
						<div className="mx-3 my-2 flex flex-col gap-3 rounded-md border border-border/50 bg-card p-4 shadow-sm">
							<div className="flex flex-col gap-1">
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
							<div className="flex items-center justify-between rounded-md bg-muted/50 px-3 py-2">
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
					) : (
						<div className="flex justify-center py-3">
							<User size={18} className="text-muted-foreground" />
						</div>
					)}
				</SidebarHeader>
				<SidebarContent>
					{state === "expanded" && (
						<SidebarGroup>
							<SidebarGroupLabel>{t("appearance")}</SidebarGroupLabel>
							<SidebarGroupContent>
								<div className="space-y-4 px-3 py-2">
									<div className="space-y-2">
										<div className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
											<Palette size={14} />
											<span>{t("theme")}</span>
										</div>
										<ThemeToggle />
									</div>
									<div className="space-y-2">
										<div className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
											<Languages size={14} />
											<span>{t("interfaceLanguage")}</span>
										</div>
										<Select
											value={language}
											onValueChange={(value) => setLanguage(value as "vi" | "en")}
										>
											<SelectTrigger className="w-full">
												<SelectValue />
											</SelectTrigger>
											<SelectContent>
												<SelectItem value="vi">{t("vietnamese")}</SelectItem>
												<SelectItem value="en">{t("english")}</SelectItem>
											</SelectContent>
										</Select>
									</div>
								</div>
							</SidebarGroupContent>
						</SidebarGroup>
					)}
				</SidebarContent>
				<SidebarFooter>
					<SidebarMenu>
						{isAdmin && (
							<SidebarMenuItem>
								<SidebarMenuButton onClick={() => setIsAdminOpen(true)}>
									<Shield />
									<span>{t("admin")}</span>
								</SidebarMenuButton>
							</SidebarMenuItem>
						)}
						<SidebarMenuItem>
							<SidebarMenuButton onClick={() => setIsSettingsOpen(true)}>
								<Settings />
								<span>{t("accountSettings")}</span>
							</SidebarMenuButton>
						</SidebarMenuItem>
						<SidebarMenuItem>
							<SidebarMenuButton onClick={handleLogout}>
								<LogOut />
								<span>{t("logout")}</span>
							</SidebarMenuButton>
						</SidebarMenuItem>
					</SidebarMenu>
				</SidebarFooter>
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
