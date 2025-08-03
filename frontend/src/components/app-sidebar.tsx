"use client";

import * as React from "react";
import { LogOut, Settings } from "lucide-react";
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
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { useAuth } from "@/contexts/auth";
import { userAPI } from "@/lib/api";

// Mock user data
const mockUser = {
	id: "mock-user-id",
	email: "user@example.com",
	name: "John Doe",
	age: 25 as number | null,
	gender: "male",
	city: "Ho Chi Minh City",
	isVisible: true,
};

export function AppSidebar({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
	const { state } = useSidebar();
	const [isSettingsOpen, setIsSettingsOpen] = React.useState(false);
	const [userData, setUserData] = React.useState(mockUser);
	const router = useRouter();
	const { logout } = useAuth();

	// Load user data from backend if available
	React.useEffect(() => {
		const loadUserData = async () => {
			try {
				// Try to load from backend first
				const response = await userAPI.getState();

				if (response.data) {
					setUserData({
						id: response.data.id,
						email: response.data.email || "unknown@example.com",
						name: response.data.name || "Unknown",
						age: response.data.profile?.age || null,
						gender:
							response.data.profile?.is_male === true
								? "male"
								: response.data.profile?.is_male === false
								? "female"
								: "other",
						city: response.data.profile?.city || "---",
						isVisible: !response.data.profile?.is_hidden,
					});
					return;
				}
			} catch (error) {
				// Backend not available, use mock data
				console.log("Backend not available, using mock data:", error);
			}
		};

		loadUserData();
	}, []);

	const handleLogout = async () => {
		await logout();
		router.push("/login");
	};

	// Helper function to display gender in Vietnamese
	const getGenderDisplay = (genderValue: string) => {
		switch (genderValue) {
			case "male":
				return "Nam";
			case "female":
				return "Nữ";
			case "other":
				return "Khác";
			default:
				return "Khác";
		}
	};

	// Handle visibility toggle in sidebar
	const handleVisibilityToggle = async (isVisible: boolean) => {
		try {
			// Update visibility in backend
			await userAPI.updateProfile({
				is_hidden: !isVisible,
			});
			// Update local state
			setUserData((prev) => ({ ...prev, isVisible }));
		} catch (error) {
			// Error already handled by toast in API client
			console.error("Visibility toggle error:", error);
		}
	};

	// Handle settings dialog save (age, gender, city)
	const handleSaveSettings = async (newSettings: { age: number | null; gender: string; city: string }) => {
		try {
			// Try to save to backend first
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
		} catch (error) {
			// Error already handled by toast in API client
			console.error("Settings save error:", error);
		}

		// Update local state with new settings
		setUserData((prev) => ({
			...prev,
			age: newSettings.age,
			gender: newSettings.gender,
			city: newSettings.city,
		}));
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
				<SidebarContent>{/* You can add more navigation items here if needed */}</SidebarContent>
				<SidebarFooter>
					<SidebarMenu>
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
