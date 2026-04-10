"use client";

import React, { useState } from "react";
import { Flag, Volume2, VolumeX, X } from "lucide-react";
import { useAuth } from "@/contexts/auth";
import { useAdmin } from "@/contexts/admin";
import { useAlertDialogContext } from "@/contexts/alert-dialog";
import { useLanguage, useTheme } from "@/contexts/theme";
import { ActionButton } from "./header/action-button";
import { Button } from "@/components/ui/button";
import { moderationAPI } from "@/lib/api";
import { toast } from "sonner";

interface HeaderProps {
    trigger: React.ReactNode;
}

export default function Header({ trigger }: HeaderProps) {
    const { room } = useAuth();
    const { isAdminOpen, setIsAdminOpen } = useAdmin();
    const { soundEnabled, toggleSound } = useTheme();
    const { t } = useLanguage();
    const partner = room?.partner;
    const [reported, setReported] = useState(false);
    const alertDialog = useAlertDialogContext();

    const roomId = room?.id;
    React.useEffect(() => {
        setReported(false);
    }, [roomId]);

    const handleReport = async () => {
        if (!partner || !room || reported) return;
        const confirmed = await alertDialog.open({
            title: t("reportDialogTitle"),
            description: t("reportDialogDescription"),
            confirmText: t("report"),
            cancelText: t("cancel"),
        });
        if (!confirmed) return;
        try {
            await moderationAPI.createReport(partner.id, room.id);
            setReported(true);
            toast.success(t("reportSubmitted"));
        } catch {
            // error toast handled by apiCall
        }
    };
    const isHidden = partner?.profile?.is_hidden;

    const partnerName = isHidden ? "Ẩn danh" : partner?.nickname || partner?.name || "Người dùng";
    const partnerAge =
        !isHidden && partner?.profile?.age
            ? `${partner.profile.age} tuổi`
            : null;
    const partnerGender =
        !isHidden &&
        partner?.profile?.is_male !== null &&
        partner?.profile?.is_male !== undefined
            ? partner.profile.is_male
                ? "Nam"
                : "Nữ"
            : null;
    const partnerSubtitle = [partnerAge, partnerGender]
        .filter(Boolean)
        .join(" • ");

    if (isAdminOpen) {
        return (
            <header className="absolute top-0 left-0 right-0 flex h-16 shrink-0 items-center justify-between border-b-2 px-4">
                <div className="flex items-center gap-2">
                    {trigger}
                    <span className="text-base font-semibold">Admin Panel</span>
                </div>
                <Button variant="secondary" size="sm" onClick={() => setIsAdminOpen(false)}>
                    <X />
                    Back to chat
                </Button>
            </header>
        );
    }

    return (
        <header className="absolute top-0 left-0 right-0 flex h-16 shrink-0 items-center justify-between border-b-2 px-4">
            <div className="flex items-center gap-2">
                {trigger}
                {partner && (
                    <div className="flex flex-col leading-tight">
                        <span className="text-base font-semibold">
                            {partnerName}
                        </span>
                        {partnerSubtitle && (
                            <span className="text-xs text-muted-foreground">
                                {partnerSubtitle}
                            </span>
                        )}
                    </div>
                )}
            </div>
            <div className="flex items-center gap-2">
                {partner && !reported && (
                    <button
                        onClick={handleReport}
                        className="w-8 h-8 flex items-center justify-center rounded-full text-muted-foreground hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950 transition-colors cursor-pointer"
                        title={t("reportUser")}
                    >
                        <Flag size={16} />
                    </button>
                )}
                <button
                    onClick={toggleSound}
                    className="w-8 h-8 flex items-center justify-center rounded-full text-muted-foreground hover:text-foreground hover:bg-muted transition-colors cursor-pointer"
                    title={soundEnabled ? t("turnOffSound") : t("turnOnSound")}
                >
                    {soundEnabled ? <Volume2 size={16} /> : <VolumeX size={16} />}
                </button>
                <ActionButton />
            </div>
        </header>
    );
}
