import { IResponse, IExtendedResponse } from './response_models';

const headers: { [id: string]: string } = {
    'Content-Type': 'application/json',
};

export async function sendRequest<T, U extends IResponse>(
    url: string,
    verb: 'put' | 'post' | 'get' | 'delete',
    path: string,
    request?: T | undefined
): Promise<U> {
    const response = await sendRequestExt<T, U>(url, verb, path, request);
    return response.body;
}

export async function sendRequestExt<T, U extends IResponse | null>(
    apiUrl: string,
    verb: 'put' | 'post' | 'get' | 'delete',
    path: string,
    request?: T | undefined
): Promise<IExtendedResponse<U>> {
    let fetchResponse: Response;

    try {
        if(!path.startsWith("/"))
            path = "/" + path;        
        const url = `${apiUrl}/${path}`;
        fetchResponse = await fetch(url, {
            method: verb,
            headers,
            body: JSON.stringify(request),
        });

        if (!fetchResponse) {
            throw new Error('No data was returned from the API');
        }

        try {
            const response = await fetchResponse.json();
            return { body: response, response: fetchResponse };
        } catch (err) {
            const error = err as Error;
            if (fetchResponse.ok) {
                throw new Error(error.message);
            } else {
                const text = await fetchResponse.text();
                throw new Error(error.message + '   ---   ' + text);
            }
        }
    } catch (err) {
        const error = err as Error;
        throw new Error('sendRequest: ' + error.message);
    }
}
