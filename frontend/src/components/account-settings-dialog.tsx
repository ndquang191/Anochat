"use client";

import * as React from "react";
import { Mars, Venus } from "lucide-react";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { useLanguage } from "@/contexts/theme";
import { cn } from "@/lib/utils";
import { MAX_AGE, MIN_AGE } from "@/types";

const selectedGenderClass =
	"border-primary bg-primary/10 text-primary hover:bg-primary/15 hover:text-primary dark:border-primary/80 dark:bg-primary/20 dark:text-primary dark:ring-1 dark:ring-primary/50 dark:hover:bg-primary/30 dark:hover:text-primary";

interface AccountSettingsData {
	nickname: string;
	birthYear: number | null;
	gender: string;
}

interface AccountSettingsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	initialData: AccountSettingsData & {
		nicknameChangeAvailableAt: number | null;
	};
	onSave: (data: AccountSettingsData) => Promise<void> | void;
}

export function AccountSettingsDialog({
	open,
	onOpenChange,
	initialData,
	onSave,
}: AccountSettingsDialogProps) {
	const { language, t } = useLanguage();
	const [nickname, setNickname] = React.useState(initialData.nickname);
	const [birthYearInput, setBirthYearInput] = React.useState(
		initialData.birthYear?.toString() ?? "",
	);
	const [gender, setGender] = React.useState(() => {
		const validGenders = ["male", "female"];
		return validGenders.includes(initialData.gender) ? initialData.gender : "male";
	});
	const [isLoading, setIsLoading] = React.useState(false);

	React.useEffect(() => {
		if (open) {
			setNickname(initialData.nickname);
			setBirthYearInput(initialData.birthYear?.toString() ?? "");
			const validGenders = ["male", "female"];
			setGender(validGenders.includes(initialData.gender) ? initialData.gender : "male");
		}
	}, [
		open,
		initialData.birthYear,
		initialData.gender,
		initialData.nickname,
		initialData.nicknameChangeAvailableAt,
	]);

	const nicknameAvailableAt = initialData.nicknameChangeAvailableAt
		? new Date(initialData.nicknameChangeAvailableAt * 1000)
		: null;
	const nicknameLocked =
		nicknameAvailableAt !== null && nicknameAvailableAt.getTime() > Date.now();
	const normalizedNickname = nickname.trim();
	const nicknameLength = Array.from(normalizedNickname).length;
	const nicknameInvalid =
		normalizedNickname !== "" && (nicknameLength < 2 || nicknameLength > 32);
	const currentYear = new Date().getFullYear();
	const minimumBirthYear = currentYear - MAX_AGE;
	const maximumBirthYear = currentYear - MIN_AGE;
	const birthYearComplete = birthYearInput.length === 4;
	const birthYear = birthYearComplete ? Number(birthYearInput) : null;
	const birthYearInvalid =
		birthYearComplete &&
		birthYear !== null &&
		(birthYear < minimumBirthYear || birthYear > maximumBirthYear);
	const birthYearIncomplete = birthYearInput.length > 0 && !birthYearComplete;
	const age =
		birthYearComplete && birthYear !== null
			? currentYear - birthYear
			: null;
	const birthYearHint =
		age !== null && age < 15
			? t("birthYearYoungHint")
			: age !== null && age > 90
				? t("birthYearOldHint")
				: t("birthYearHint");

	const handleSave = async () => {
		setIsLoading(true);

		try {
			await onSave({
				nickname: normalizedNickname,
				birthYear,
				gender,
			});
			toast.success(t("userInfoSaveSuccess"));
			onOpenChange(false);
		} catch (error) {
			console.error("Error saving user settings:", error);
			toast.error(t("somethingWentWrong"));
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="z-[9998] h-fit max-h-[95vh] overflow-y-auto sm:max-w-[460px]">
				<DialogHeader>
					<DialogTitle>{t("accountSettings")}</DialogTitle>
					<DialogDescription>
						{initialData.birthYear === null
							? t("completeProfileHint")
							: t("accountSettingsDescription")}
					</DialogDescription>
				</DialogHeader>
				<div className="grid gap-4 py-1">
					<div className="grid grid-cols-4 items-start gap-4">
						<Label htmlFor="display-name" className="pt-2 text-right">
							{t("displayName")}
						</Label>
						<div className="col-span-3 space-y-1.5">
							<Input
								id="display-name"
								value={nickname}
								onChange={(event) => setNickname(event.target.value)}
								placeholder={t("displayNamePlaceholder")}
								disabled={nicknameLocked || isLoading}
								maxLength={32}
							/>
							<p className="text-xs text-muted-foreground">
								{nicknameLocked && nicknameAvailableAt
									? t("displayNameLockedUntil", {
											date: nicknameAvailableAt.toLocaleDateString(
												language === "vi" ? "vi-VN" : "en-US"
											),
										})
									: t("displayNameHint")}
							</p>
						</div>
					</div>
					<div className="grid grid-cols-4 items-start gap-4">
						<Label htmlFor="birth-year" className="pt-2 text-right">
							{t("birthYear")}
						</Label>
						<div className="col-span-3 space-y-1.5">
							<Input
								id="birth-year"
								type="text"
								inputMode="numeric"
								pattern="[0-9]{4}"
								maxLength={4}
								value={birthYearInput}
								onChange={(e) => {
									const value = e.target.value.replace(/\D/g, "").slice(0, 4);
									setBirthYearInput(value);
								}}
								placeholder={t("birthYearPlaceholder")}
							/>
							<p className="text-xs text-muted-foreground">{birthYearHint}</p>
						</div>
					</div>
					<div className="grid grid-cols-4 items-center gap-4">
						<Label className="text-right">{t("gender")}</Label>
						<div className="col-span-3 flex gap-3">
							<Button
								type="button"
								variant="outline"
								size="icon"
								onClick={() => setGender("male")}
								aria-label={t("male")}
								aria-pressed={gender === "male"}
								title={t("male")}
								className={cn(
									"h-10 w-10",
									gender === "male" && selectedGenderClass
								)}
							>
								<Mars className="size-5" />
							</Button>
							<Button
								type="button"
								variant="outline"
								size="icon"
								onClick={() => setGender("female")}
								aria-label={t("female")}
								aria-pressed={gender === "female"}
								title={t("female")}
								className={cn(
									"h-10 w-10",
									gender === "female" && selectedGenderClass
								)}
							>
								<Venus className="size-5" />
							</Button>
						</div>
					</div>
				</div>
				<DialogFooter>
					<Button
						variant="outline"
						onClick={() => onOpenChange(false)}
						disabled={isLoading}
					>
						{t("cancel")}
					</Button>
					<Button
						onClick={handleSave}
						disabled={
							isLoading ||
							nicknameInvalid ||
							birthYearInvalid ||
							birthYearIncomplete
						}
					>
						{isLoading ? t("saving") : t("saveChanges")}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
