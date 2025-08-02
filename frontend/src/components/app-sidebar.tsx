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

// Mock user data
const mockUser = {
	id: "mock-user-id",
	email: "user@example.com",
	name: "John Doe",
	nickname: "Johnny",
	gender: "male",
	address: "Ho Chi Minh City",
	isVisible: true,
};

export function AppSidebar({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
	const { state } = useSidebar();
	const [isSettingsOpen, setIsSettingsOpen] = React.useState(false);
	const [userData, setUserData] = React.useState(mockUser);
	const router = useRouter();

	const handleLogout = async () => {
		// Placeholder for future logout implementation
		console.log("Logout functionality will be implemented here");
		router.push("/login");
	};

	// Helper function to display gender in Vietnamese
	const getGenderDisplay = (genderValue: string) => {
		switch (genderValue) {
			case 'male': return 'Nam';
			case 'female': return 'Nữ';
			case 'other': return 'Khác';
			default: return 'Khác';
		}
	};

	const handleSaveSettings = async (newSettings: typeof userData) => {
		// Update local state
		setUserData(newSettings);
		setIsSettingsOpen(false);
		
		// Placeholder for future settings save implementation
		console.log("Settings save functionality will be implemented here", newSettings);
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
											{userData.address !== "---" ? `${getGenderDisplay(userData.gender)}, ${userData.address}` : getGenderDisplay(userData.gender)}
										</span>
										<div className="flex items-center gap-2 mt-2">
											<Switch
												id="public-status"
												checked={userData.isVisible}
												onCheckedChange={(checked) =>
													setUserData((prev) => ({ ...prev, isVisible: checked }))
												}
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
				initialData={userData}
				onSave={handleSaveSettings}
			/>
		</>
	);
}
