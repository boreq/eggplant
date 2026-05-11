import { Component, Vue } from 'vue-property-decorator';

@Component
export default class Notifications extends Vue {

    static readonly notificationEvent = 'eggplant_notification';

    static pushError(vue: Vue, text: string, error?: any): void {
        const extra = error && error.response && error.response.data
            && error.response.data.message ? error.response.data.message : null;

        const notification: Notification = {
            id: this.notificationId++,
            class: 'error',
            created: new Date(),
            text: text,
            extra: extra,
        };
        vue.$root.$emit(this.notificationEvent, notification);
    }

    static pushSuccess(vue: Vue, text: string): void {
        const notification: Notification = {
            id: this.notificationId++,
            class: 'success',
            created: new Date(),
            text: text,
            extra: null,
        };
        vue.$root.$emit(this.notificationEvent, notification);
    }

    private static notificationId = 0;
    private static readonly visibilityDuration = 10;
    private static readonly animationDuration = 2;

    notifications: Notification[] = [];

    private intervalID: number;

    mounted(): void {
        this.$root.$on(Notifications.notificationEvent, (notification: Notification) => {
            this.notifications.splice(0, 0, notification);
        });
        this.intervalID = window.setInterval(this.processErrors, 100);
    }

    destroyed(): void {
        window.clearInterval(this.intervalID);
    }

    shouldHide(notification: Notification): boolean {
        const secondsDuration = this.duration(new Date(), notification.created);
        return secondsDuration > Notifications.visibilityDuration;
    }

    private processErrors(): void {
        this.notifications = this.notifications.filter(notification => {
            const secondsDuration = this.duration(new Date(), notification.created);
            return secondsDuration < (Notifications.visibilityDuration + Notifications.animationDuration);
        });
    }

    private duration(a: Date, b: Date): number {
        return (a.getTime() - b.getTime()) / 1000;
    }
}

class Notification {
    id: number;
    class: string;
    created: Date;
    text: string;
    extra: string;
}
