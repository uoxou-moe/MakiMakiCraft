import { SSMClient, SendCommandCommand } from "@aws-sdk/client-ssm";

const ssmClient = new SSMClient({});

export const handler = async (event) => {
    console.log("Step 1: Attempting to send SSM command...");

    // 存在しない、ダミーのインスタンスIDを定義
    const instanceId = 'i-dummy';

    // SendCommand APIに渡すパラメータを準備
    const params = {
        DocumentName: "AWS-RunShellScript",
        InstanceIds: [instanceId],
        Parameters: {
            // 実行するコマンド（この時点では何でも良い）
            commands: ["echo 'This command will fail due to permissions.'"]
        },
    };

    try {
        // コマンド送信を試みる
        const command = new SendCommandCommand(params);
        const response = await ssmClient.send(command);

        console.log("Command sent successfully (This should not happen yet):", response);
        return {
            statusCode: 200,
            body: JSON.stringify('Command sent successfully.'),
        };

    } catch (error) {
        // エラーを捕捉してログに出力する（ここに来るのが期待される動作）
        console.error("Caught expected error:", error);

        return {
            statusCode: 500,
            body: JSON.stringify({
                message: "Failed as expected.",
                errorName: error.name,
                errorMessage: error.message
            }),
        };
    }
};
