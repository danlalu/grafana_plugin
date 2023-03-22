export declare const promLanguageDefinition: {
    id: string;
    extensions: string[];
    aliases: string[];
    mimetypes: any[];
    loader: () => Promise<typeof import("./promql")>;
};
