// Client-side cookie utilities
export const setCookie = (name: string, value: string, days: number = 7) => {
	const expires = new Date();
	expires.setTime(expires.getTime() + days * 24 * 60 * 60 * 1000);
	const cookieString = `${name}=${value};expires=${expires.toUTCString()};path=/;SameSite=Lax`;
	document.cookie = cookieString;
};

export const getCookie = (name: string): string | null => {
	const value = `; ${document.cookie}`;
	const parts = value.split(`; ${name}=`);
	if (parts.length === 2) return parts.pop()?.split(";").shift() || null;
	return null;
};

export const deleteCookie = (name: string) => {
	document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/;`;
};
