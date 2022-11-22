export enum SetupType {
    HOSTNAME = "hostname",
    INPUT = "input",
    SECRET = "secret",
    TEXTBOX = "textbox",
    BOOLEAN = "boolean",
}

export type SetupOption = {
    id: string;
    type: SetupType;
    name: string;
    description: string;
    sticky: boolean;
    must_secure?: boolean;
    regexp?: string;
    required: boolean;
};

export class POError extends Error {}
