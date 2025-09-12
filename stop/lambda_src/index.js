import { SSMClient, SendCommandCommand } from "@aws-sdk/client-ssm";

const ssmClient = new SSMClient({});

export const handler = async (event) => {
    console.log("Step 1: Attempting to send SSM command...");

    const instanceId = process.env.TARGET_INSTANCE_ID;

    if (!instanceId) {
        console.error("Error: TARGET_INSTANCE_ID environment variable is not set.");
        return {
            statusCode: 500,
            body: JSON.stringify({
                message: "TARGET_INSTANCE_ID is not configured.",
            }),
        };
    }

    console.log(`Target instance ID: ${instanceId}`);

    // create file that now date written at /tmp dir
    const command = "date > /tmp/lambda_ssm_test.txt";

    // SendCommand APIに渡すパラメータを準備
    const params = {
        DocumentName: "AWS-RunShellScript",
        InstanceIds: [instanceId],
        Parameters: {
            commands: [command]
        },
        Comment: `Command from Lambda at ${new Date().toISOString()}`
    };

    try {
        const commandRequest = new SendCommandCommand(params);
        const response = await ssmClient.send(commandRequest);
        const commandId = response.Command.CommandId;

        console.log("Command sent successfully Command ID:", commandId);
        return {
            statusCode: 200,
            body: JSON.stringify({
                message: "Command sent successfully.",
                commandId: commandId
            }),
        };
    } catch (error) {
        console.error("Error sending command:", error);

        return {
            statusCode: 500,
            body: JSON.stringify({
                message: "Failed to send command.",
                errorName: error.name,
                errorMessage: error.message
            }),
        };
    }
};
