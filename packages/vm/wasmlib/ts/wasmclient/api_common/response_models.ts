export interface IResponse {
    error?: string;
}

export interface IExtendedResponse<U> {
    body: U;
    response: Response;
}
