"use client";

import * as React from "react";
import { LogOut, Settings, Shield } from "lucide-react";
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
import { UserSettingsDialog } from "@/components/user-settings-dialog";
import { AdminPanel } from "@/components/admin/admin-panel";
import { AdminUserID } from "@/types";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { useAuth } from "@/contexts/auth";
import { userAPI } from "@/lib/api";
import { useUserState, useInvalidateUserState } from "@/hooks/queries/use-user-state";

interface UserData {
	id: string;
	email: string;
	name: string;
	age: number | null;
	gender: string;
	city: string;
	isVisible: boolean;
}

const defaultUserData: UserData = {
	id: "",
	email: "",
	name: "...",
	age: null,
	gender: "other",
	city: "---",
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
		city: profile?.city || "---",
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
	const { data } = useUserState();
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

	const handleSaveSettings = async (newSettings: { age: number | null; gender: string; city: string }) => {
		try {
			await userAPI.updateProfile({
				age: newSettings.age,
				city: newSettings.city,
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
			city: newSettings.city,
		}));
		invalidateUserState();
		setIsSettingsOpen(false);
	};

	return (
		<>
			<Sidebar className={cn(className)} {...props}>
				<SidebarHeader>
					<SidebarGroup>
						<SidebarGroupContent className="flex flex-col items-start gap-2 p-4">
							<div className="flex flex-col items-start">
								<span className="text-lg font-semibold whitespace-nowrap overflow-hidden text-ellipsis max-w-full">
									{userData.name}
								</span>
								{state === "expanded" && (
									<>
										<span className="text-sm text-muted-foreground">
											{userData.city !== "---"
												? `${getGenderDisplay(userData.gender)}, ${userData.city}`
												: getGenderDisplay(userData.gender)}
										</span>
										<div className="flex items-center gap-2 mt-2">
											<Switch
												id="public-status"
												checked={userData.isVisible}
												onCheckedChange={handleVisibilityToggle}
											/>
											<Label htmlFor="public-status" className="text-sm">
												{userData.isVisible ? "Công khai thông tin" : "Riêng tư"}
											</Label>
										</div>
									</>
								)}
							</div>
						</SidebarGroupContent>
					</SidebarGroup>
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
					city: userData.city,
				}}
				onSave={handleSaveSettings}
			/>
		</>
	);
}
