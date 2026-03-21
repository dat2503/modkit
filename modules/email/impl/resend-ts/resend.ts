import type { IEmailService, EmailMessage, EmailResult } from '../../../contracts/ts/email'

export interface ResendConfig {
  apiKey: string
  fromDefault: string
}

/**
 * ResendEmailService implements IEmailService using Resend (resend.com).
 */
export class ResendEmailService implements IEmailService {
  constructor(private readonly config: ResendConfig) {}

  async send(msg: EmailMessage): Promise<EmailResult> {
    // TODO: implement using resend npm package
    // const resend = new Resend(this.config.apiKey)
    // const { data, error } = await resend.emails.send({...})
    throw new Error('not implemented')
  }

  async sendBatch(msgs: EmailMessage[]): Promise<EmailResult[]> {
    // TODO: implement using resend.emails.sendBatch()
    throw new Error('not implemented')
  }
}
