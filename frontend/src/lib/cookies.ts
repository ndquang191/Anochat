// Client-side cookie utilities
export const setCookie = (name: string, value: string, days: number = 7) => {
	const expires = new Date();
	expires.setTime(expires.getTime() + days * 24 * 60 * 60 * 1000);
	const cookieString = `${name}=${value};expires=${expires.toUTCString()};path=/;SameSite=Lax`;
	console.log(`Setting cookie: ${name}=${value.substring(0, 20)}...`);
	console.log(`Cookie string: ${cookieString}`);
	document.cookie = cookieString;
	console.log(`Cookies after setting ${name}:`, document.cookie);
};

export const getCookie = (name: string): string | null => {
	const value = `; ${document.cookie}`;
	const parts = value.split(`; ${name}=`);
	if (parts.length === 2) return parts.pop()?.split(";").shift() || null;
	return null;
};

export const deleteCookie = (name: string) => {
	console.log(`Deleting cookie: ${name}`);
	document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/;`;
	console.log(`Cookie ${name} deleted, current cookies:`, document.cookie);
};
