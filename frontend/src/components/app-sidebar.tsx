"use client";

import * as React from "react";
import { LogOut, Settings, Shield, User } from "lucide-react";
import {
	Sidebar,
	SidebarContent,
	SidebarFooter,
	SidebarGroup,
	SidebarGroupContent,
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
import { UserSettingsDialog } from "@/components/user-settings-dialog";
import { AdminPanel } from "@/components/admin/admin-panel";
import { AdminUserID } from "@/types";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { AuroraText } from "@/components/aurora-text";
import { useAuth } from "@/contexts/auth";
import { userAPI } from "@/lib/api";
import { useUserState, useInvalidateUserState } from "@/hooks/queries/use-user-state";

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

function deriveUserData(user: ReturnType<typeof useAuth>["user"], data: ReturnType<typeof useUserState>["data"]): UserData {
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

export function AppSidebar({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
	const { state } = useSidebar();
	const [isSettingsOpen, setIsSettingsOpen] = React.useState(false);
	const [isAdminOpen, setIsAdminOpen] = React.useState(false);
	const [localOverrides, setLocalOverrides] = React.useState<Partial<UserData>>({});
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
				return "Nam";
			case "female":
				return "Nữ";
			default:
				return "Khác";
		}
	};

	const handleVisibilityToggle = async (isVisible: boolean) => {
		try {
			await userAPI.updateProfile({
				is_hidden: !isVisible,
			});
			setLocalOverrides((prev) => ({ ...prev, isVisible }));
			invalidateUserState();
		} catch {
		}
	};

	const handleSaveSettings = async (newSettings: { age: number | null; gender: string }) => {
		try {
			await userAPI.updateProfile({
				age: newSettings.age,
				is_male:
					newSettings.gender === "male"
						? true
						: newSettings.gender === "female"
						? false
						: undefined,
			});
		} catch {
		}

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
				<div className="px-6 pt-4 pb-2" style={{ fontFamily: "var(--font-changa-one)" }}>
					<AuroraText className="text-4xl tracking-widest">
						ANOCHAT
					</AuroraText>
				</div>
			<SidebarHeader>
					{state === "expanded" ? (
						<div className="mx-3 my-2 rounded-md bg-card border border-border/50 shadow-sm p-4 flex flex-col gap-3">
							<div className="flex flex-col gap-1">
								{isLoading && !data ? (
									<Skeleton className="h-5 w-28 rounded" />
								) : (
									<span className="text-base font-semibold truncate leading-tight">
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
											{userData.age ? `, ${userData.age} tuổi` : ""}
										</span>
									)}
								</div>
							</div>
							<div className="flex items-center justify-between rounded-md bg-muted/50 px-3 py-2">
								<Label htmlFor="public-status" className="text-xs text-muted-foreground cursor-pointer select-none">
									{userData.isVisible ? "Công khai thông tin" : "Riêng tư"}
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
				<SidebarContent />
				<SidebarFooter>
					<SidebarMenu>
						{isAdmin && (
							<SidebarMenuItem>
								<SidebarMenuButton onClick={() => setIsAdminOpen(true)}>
									<Shield />
									<span>Admin</span>
								</SidebarMenuButton>
							</SidebarMenuItem>
						)}
						<SidebarMenuItem>
							<SidebarMenuButton onClick={() => setIsSettingsOpen(true)}>
								<Settings />
								<span>Cài đặt</span>
							</SidebarMenuButton>
						</SidebarMenuItem>
						<SidebarMenuItem>
							<SidebarMenuButton onClick={handleLogout}>
								<LogOut />
								<span>Đăng xuất</span>
							</SidebarMenuButton>
						</SidebarMenuItem>
					</SidebarMenu>
				</SidebarFooter>
				<SidebarRail />
			</Sidebar>

			{isAdmin && <AdminPanel open={isAdminOpen} onOpenChange={setIsAdminOpen} />}
			<UserSettingsDialog
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
