"use client";

import * as React from "react";
import { LogOut, Settings, Shield, User } from "lucide-react";
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
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
import { AdminUserID } from "@/types";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { AuroraText } from "@/components/aurora-text";
import { useAuth } from "@/contexts/auth";
import { useAdmin } from "@/contexts/admin";
import { userAPI } from "@/lib/api";
import {
    useUserState,
    useInvalidateUserState,
} from "@/hooks/queries/use-user-state";
import { toast } from "sonner";

interface UserData {
    id: string;
    email: string;
    name: string;
    nickname: string;
    age: number | null;
    gender: string;
    isVisible: boolean;
}

const defaultUserData: UserData = {
    id: "",
    email: "",
    name: "...",
    nickname: "",
    age: null,
    gender: "other",
    isVisible: false,
};

function deriveUserData(
    user: ReturnType<typeof useAuth>["user"],
    data: ReturnType<typeof useUserState>["data"],
): UserData {
    if (!user) return defaultUserData;
    const profile = data?.profile;
    return {
        id: user.id,
        email: user.email || "",
        name: user.name || "Unknown",
        nickname: profile?.nickname || "",
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

export function AppSidebar({
    className,
    ...props
}: React.HTMLAttributes<HTMLDivElement>) {
    const { state } = useSidebar();
    const [isSettingsOpen, setIsSettingsOpen] = React.useState(false);
    const { setIsAdminOpen } = useAdmin();
    const [localOverrides, setLocalOverrides] = React.useState<
        Partial<UserData>
    >({});
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
            toast.error("Không thể cập nhật trạng thái hiển thị");
        }
    };

    const handleSaveSettings = async (newSettings: {
        nickname: string;
        birthYear: number | null;
        gender: string;
    }) => {
        const currentYear = new Date().getFullYear();
        const age = newSettings.birthYear
            ? currentYear - newSettings.birthYear
            : null;
        try {
            await userAPI.updateProfile({
                nickname: newSettings.nickname || null,
                age,
                is_male:
                    newSettings.gender === "male"
                        ? true
                        : newSettings.gender === "female"
                          ? false
                          : undefined,
            });
        } catch {
            toast.error("Không thể lưu cài đặt");
        }

        setLocalOverrides((prev) => ({
            ...prev,
            nickname: newSettings.nickname,
            age,
            gender: newSettings.gender,
        }));
        invalidateUserState();
        setIsSettingsOpen(false);
    };

    return (
        <>
            <Sidebar className={cn(className)} {...props}>
                <div
                    className="px-6 pt-4 pb-2"
                    style={{ fontFamily: "var(--font-changa-one)" }}
                >
                    <AuroraText className="text-4xl tracking-widest">
                        ANOCHAT
                    </AuroraText>
                </div>
                <SidebarHeader>
                    {state === "expanded" ? (
                        <div className="mx-3 my-2 rounded-md bg-card border border-border/50 shadow-sm overflow-hidden">
                            {/* Top row: name + settings */}
                            <div className="flex items-center gap-2 px-3 pt-3 pb-2">
                                <div className="flex-1 min-w-0">
                                    {isLoading && !data ? (
                                        <Skeleton className="h-4 w-28 rounded mb-1" />
                                    ) : (
                                        <p className="text-sm font-semibold truncate leading-tight">
                                            {userData.nickname || userData.name}
                                        </p>
                                    )}
                                    {isLoading && !data ? (
                                        <Skeleton className="h-3 w-16 rounded" />
                                    ) : (
                                        <p className="text-xs text-muted-foreground truncate">
                                            {getGenderDisplay(userData.gender)}
                                            {userData.age ? `, ${userData.age} tuổi` : ""}
                                        </p>
                                    )}
                                </div>
                                <button
                                    onClick={() => setIsSettingsOpen(true)}
                                    className="shrink-0 p-1.5 rounded-md hover:bg-muted transition-colors text-muted-foreground hover:text-foreground cursor-pointer"
                                    title="Cài đặt"
                                >
                                    <Settings className="h-4 w-4" />
                                </button>
                            </div>
                            {/* Bottom row: visibility toggle */}
                            <div className="flex items-center justify-between border-t border-border/40 bg-muted/30 px-3 py-2">
                                {isLoading && !data ? (
                                    <>
                                        <Skeleton className="h-3 w-24 rounded" />
                                        <Skeleton className="h-5 w-9 rounded-full" />
                                    </>
                                ) : (
                                    <>
                                        <Label
                                            htmlFor="public-status"
                                            className="text-xs text-muted-foreground cursor-pointer select-none"
                                        >
                                            {userData.isVisible ? "Công khai" : "Ẩn danh"}
                                        </Label>
                                        <Switch
                                            id="public-status"
                                            checked={userData.isVisible}
                                            onCheckedChange={handleVisibilityToggle}
                                        />
                                    </>
                                )}
                            </div>
                        </div>
                    ) : (
                        <div className="flex justify-center py-3">
                            <User className="h-4 w-4 text-muted-foreground" />
                        </div>
                    )}
                </SidebarHeader>
                <SidebarContent />
                <SidebarFooter>
                    <SidebarMenu>
                        {isAdmin && (
                            <SidebarMenuItem>
                                <SidebarMenuButton
                                    onClick={() => setIsAdminOpen(true)}
                                >
                                    <Shield size={20} />
                                    <span>Admin</span>
                                </SidebarMenuButton>
                            </SidebarMenuItem>
                        )}
                        <SidebarMenuItem>
                            <SidebarMenuButton
                                onClick={handleLogout}
                                className="text-sm md:text-base"
                            >
                                <LogOut size={20} />
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
                    nickname: userData.nickname,
                    birthYear: userData.age
                        ? new Date().getFullYear() - userData.age
                        : null,
                    gender: userData.gender,
                }}
                onSave={handleSaveSettings}
            />
        </>
    );
}
