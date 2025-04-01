import chalk from "chalk";

export class Logger {
    static info(message: string): void {
      console.log(chalk.white(message));
    }
  
    static success(message: string): void {
      console.log(chalk.green(`✓ ${message}`));
    }
  
    static warn(message: string): void {
      console.log(chalk.yellow(`⚠ ${message}`));
    }
  
    static error(message: string): void {
      console.log(chalk.red(`✗ ${message}`));
    }
  
    static header(message: string): void {
      console.log(chalk.cyan(`\n${message}\n`));
    }
  
    static divider(): void {
      console.log('\n---------------------------------------------------------\n');
    }
  }