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

const selectedGenderClass =
	"border-primary bg-primary/10 text-primary hover:bg-primary/15 hover:text-primary dark:border-primary/80 dark:bg-primary/20 dark:text-primary dark:ring-1 dark:ring-primary/50 dark:hover:bg-primary/30 dark:hover:text-primary";

interface AccountSettingsData {
	nickname: string;
	age: number | null;
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
	const [age, setAge] = React.useState(initialData.age);
	const [gender, setGender] = React.useState(() => {
		const validGenders = ["male", "female"];
		return validGenders.includes(initialData.gender) ? initialData.gender : "male";
	});
	const [isLoading, setIsLoading] = React.useState(false);

	React.useEffect(() => {
		if (open) {
			setNickname(initialData.nickname);
			setAge(initialData.age);
			const validGenders = ["male", "female"];
			setGender(validGenders.includes(initialData.gender) ? initialData.gender : "male");
		}
	}, [
		open,
		initialData.age,
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

	const handleSave = async () => {
		setIsLoading(true);

		try {
			await onSave({
				nickname: normalizedNickname,
				age,
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
					<DialogDescription>{t("accountSettingsDescription")}</DialogDescription>
				</DialogHeader>
				<div className="grid gap-4 py-4">
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
					<div className="grid grid-cols-4 items-center gap-4">
						<Label htmlFor="age" className="text-right">
							{t("age")}
						</Label>
						<Input
							id="age"
							type="number"
							value={age || ""}
							onChange={(e) =>
								setAge(e.target.value ? parseInt(e.target.value, 10) : null)
							}
							className="col-span-3"
							placeholder={t("agePlaceholder")}
							min="1"
							max="120"
						/>
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
					<Button onClick={handleSave} disabled={isLoading || nicknameInvalid}>
						{isLoading ? t("saving") : t("saveChanges")}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
