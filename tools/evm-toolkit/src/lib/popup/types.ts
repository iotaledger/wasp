export type PopupProps = Record<string, unknown>

export type PopupAction = {
    title: string
    action: (() => void) | (() => Promise<void>)
    danger?: boolean
}
