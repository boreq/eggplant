import { Component, Vue, Ref } from 'vue-property-decorator';


@Component
export default class Dropdown extends Vue {

    isOpen = false;

    top = '';
    left = '';
    arrowTop = '';
    arrowLeft = '';

    @Ref('trigger')
    trigger: HTMLElement;

    created(): void {
        window.addEventListener('resize', this.onResize);
    }

    destroyed(): void {
        window.removeEventListener('resize', this.onResize);
    }

    onResize(): void {
        this.refreshPositions();
    }

    open(): void {
        this.refreshPositions();
        this.isOpen = true;
    }

    close() {
        this.isOpen = false;
    }

    private refreshPositions(): void {
        const rect = this.trigger.getBoundingClientRect();
        this.top = rect.bottom + 'px';
        this.left = rect.x + 'px';
        this.arrowTop = rect.bottom + 'px';
        this.arrowLeft = (rect.x + (rect.width / 2)) + 'px';
    }
}
