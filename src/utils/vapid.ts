import webpush from 'web-push';

export default async (vapid: { publicKey: string, privateKey: string, email: string }) => {
    webpush.setVapidDetails('mailto:' + vapid.email, vapid.publicKey, vapid.privateKey);
}