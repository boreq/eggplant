import Cookies from 'js-cookie';

const tokenCookieName = 'auth-token';

export class AuthService {

    storeToken(token: string): void {
        Cookies.set(tokenCookieName, token);
    }

    clearToken(): void {
        Cookies.remove(tokenCookieName);
    }

    getToken(): string {
        return Cookies.get(tokenCookieName);
    }

}
