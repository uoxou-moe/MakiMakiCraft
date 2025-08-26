# EC2インスタンス用のIAMロール
resource "aws_iam_role" "ec2_ssm_role" {
  name = "terraform-ec2-ssm-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

# SSMManagedInstanceCore
resource "aws_iam_role_policy_attachment" "ec2_ssm_core_attachment" {
  role       = aws_iam_role.ec2_ssm_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# AmazonEC2ReadOnlyAccess
resource "aws_iam_role_policy_attachment" "ec2_readonly_attachment" {
  role       = aws_iam_role.ec2_ssm_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess"
}

# Customer managed policy for Minecraft backup
resource "aws_iam_policy" "minecraft_backup_policy" {
  name        = "terraform-MinecraftBackupS3UploadPolicy"
  description = "Allows EC2 instances to upload Minecraft backups to specific S3 bucket."
  policy      = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
        ]
        Resource = "arn:aws:s3:::supurazako-minecraft-backup/*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "minecraft_backup_attachment" {
  role       = aws_iam_role.ec2_ssm_role.name
  policy_arn = aws_iam_policy.minecraft_backup_policy.arn
}

# EC2用のインスタンスプロファイル
resource "aws_iam_instance_profile" "ec2_ssm_instance_profile" {
  name = "terraform-ec2-ssm-instance-profile"
  role = aws_iam_role.ec2_ssm_role.name
}

# 既存EC2インスタンスを参照するデータソース
data "aws_instance" "minecraft_server" {
  filter {
    name   = "tag:Name"
    values = ["makimaki-craft"]
  }
}
